#!/usr/bin/env python3
"""User Management Service Container Cleanup.

This script cleans up the user-management-service deployment from Kubernetes. It
provides a rich terminal interface with confirmation prompts.
"""

import sys
from contextlib import suppress
from dataclasses import dataclass
from pathlib import Path

import click
import docker
from kubernetes import client, config
from kubipy.utils import minipy
from rich.console import Console
from rich.panel import Panel
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.prompt import Confirm

from .minikube_utils import check_minikube_status, start_minikube

# Configure rich console
console = Console()


@dataclass
class CleanupConfig:
    """Configuration for the cleanup process."""

    namespace: str = "user-management"
    image_name: str = "user-management-service"
    image_tag: str = "latest"
    config_dir: str = "k8s"

    @property
    def full_image_name(self) -> str:
        """Get the full Docker image name with tag."""
        return f"{self.image_name}:{self.image_tag}"


class CleanupManager:
    """Manages the cleanup process."""

    def __init__(self, config_obj: CleanupConfig) -> None:
        """Initialize the cleanup manager."""
        self.config = config_obj
        self.api_client: client.ApiClient | None = None
        self.core_v1_api: client.CoreV1Api | None = None
        self.apps_v1_api: client.AppsV1Api | None = None

    def initialize_client(self) -> bool:
        """Initialize Kubernetes client."""
        try:
            config.load_kube_config()
            self.api_client = client.ApiClient()
            self.core_v1_api = client.CoreV1Api(self.api_client)
            self.apps_v1_api = client.AppsV1Api(self.api_client)
        except Exception as e:
            console.print(
                f"❌ Failed to initialize Kubernetes client: {e}", style="red"
            )
            return False
        else:
            return True

    def start_minikube_if_needed(self) -> bool:
        """Start Minikube if not running."""
        if check_minikube_status():
            console.print("✅ Minikube is already running", style="green")
            return True

        console.print(
            "⚠️ Minikube is not running. Starting Minikube...",
            style="yellow",
        )
        try:
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console,
            ) as progress:
                task = progress.add_task("Starting Minikube...", total=None)

                if start_minikube():
                    progress.update(task, completed=True)
                    console.print("✅ Minikube started", style="green")
                    return True
                console.print("❌ Failed to start Minikube", style="red")
                return False
        except Exception as e:
            console.print(f"❌ Failed to start Minikube: {e}", style="red")
            return False

    def delete_kubernetes_resources(self) -> bool:
        """Delete Kubernetes resources."""
        try:
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console,
            ) as progress:
                task = progress.add_task("Deleting Kubernetes resources...", total=None)

                # Delete resources in order
                resources = [
                    "configmap-template.yaml",
                    "secret-template.yaml",
                    "deployment.yaml",
                    "service.yaml",
                    "ingress.yaml",
                ]

                for resource in resources:
                    resource_path = Path(self.config.config_dir) / resource
                    if resource_path.exists():
                        try:
                            # Use kubipy to delete resources
                            mk = minipy()
                            mk.kubectl(
                                [
                                    "delete",
                                    "-f",
                                    str(resource_path),
                                    "-n",
                                    self.config.namespace,
                                    "--ignore-not-found",
                                ]
                            )
                            console.print(f"✅ Deleted {resource}", style="green")
                        except Exception:
                            console.print(
                                f"⚠️ Failed to delete {resource}", style="yellow"
                            )

                progress.update(task, completed=True)
                return True
        except Exception as e:
            console.print(f"❌ Failed to delete Kubernetes resources: {e}", style="red")
            return False

    def remove_docker_image(self) -> bool:
        """Remove Docker image."""
        try:
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console,
            ) as progress:
                task = progress.add_task("Removing Docker image...", total=None)

                # Set minikube docker environment
                mk = minipy()
                mk.docker_env()

                # Remove the image using docker-py
                client = docker.from_env()
                with suppress(Exception):
                    client.images.remove(self.config.full_image_name, force=True)

                progress.update(task, completed=True)
                console.print("✅ Docker image removed", style="green")
                return True
        except Exception as e:
            console.print(f"❌ Failed to remove Docker image: {e}", style="red")
            return False

    def delete_persistent_volumes(self) -> bool:
        """Delete PersistentVolumeClaims and PersistentVolumes."""
        try:
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console,
            ) as progress:
                task = progress.add_task("Deleting persistent volumes...", total=None)

                # Delete PVCs
                mk = minipy()
                mk.kubectl(
                    [
                        "delete",
                        "-f",
                        f"{self.config.config_dir}/pvc.yaml",
                        "-n",
                        self.config.namespace,
                        "--ignore-not-found",
                    ]
                )

                # Delete PVs with app label
                mk.kubectl(
                    [
                        "delete",
                        "pv",
                        "-l",
                        "app=user-management",
                        "--ignore-not-found",
                    ]
                )

                progress.update(task, completed=True)
                console.print("✅ Persistent volumes deleted", style="green")
                return True
        except Exception as e:
            console.print(f"❌ Failed to delete persistent volumes: {e}", style="red")
            return False

    def stop_minikube(self) -> bool:
        """Stop Minikube."""
        try:
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console,
            ) as progress:
                task = progress.add_task("Stopping Minikube...", total=None)

                mk = minipy()
                mk.stop()

                progress.update(task, completed=True)
                console.print("✅ Minikube stopped", style="green")
                return True
        except Exception as e:
            console.print(f"❌ Failed to stop Minikube: {e}", style="red")
            return False


@click.command()
@click.option("--namespace", default="user-management", help="Kubernetes namespace")
@click.option("--config-dir", default="k8s", help="Kubernetes config directory")
@click.option("--force", is_flag=True, help="Skip confirmation prompts")
@click.option("--stop-minikube", is_flag=True, help="Stop Minikube after cleanup")
def main(namespace: str, config_dir: str, force: bool, stop_minikube: bool) -> None:
    """Clean up the user-management-service deployment."""
    # Initialize configuration
    config_obj = CleanupConfig(
        namespace=namespace,
        config_dir=config_dir,
    )

    # Initialize cleanup manager
    cleanup_manager = CleanupManager(config_obj)
    if not cleanup_manager.initialize_client():
        console.print("❌ Failed to initialize Kubernetes client", style="red")
        sys.exit(1)

    # Check and start Minikube if needed
    console.print(Panel("🔍 Checking Minikube Status", style="blue"))
    if not cleanup_manager.start_minikube_if_needed():
        console.print("❌ Failed to start Minikube", style="red")
        sys.exit(1)

    # Delete Kubernetes resources
    console.print(Panel("🧹 Deleting Kubernetes Resources", style="blue"))
    if not cleanup_manager.delete_kubernetes_resources():
        console.print("❌ Failed to delete Kubernetes resources", style="red")
        sys.exit(1)

    # Ask about PVC deletion
    console.print(Panel("💾 Persistent Volume Claims", style="blue"))
    if force or Confirm.ask(
        "⚠️ Do you want to delete the PersistentVolumeClaim (PVC)? "
        "This will delete all stored database data!"
    ):
        if not cleanup_manager.delete_persistent_volumes():
            console.print("❌ Failed to delete persistent volumes", style="red")
            sys.exit(1)
    else:
        console.print("💾 PVC retained", style="green")

    # Remove Docker image
    console.print(Panel("🐳 Removing Docker Image", style="blue"))
    if not cleanup_manager.remove_docker_image():
        console.print("❌ Failed to remove Docker image", style="red")
        sys.exit(1)

    # Ask about stopping Minikube
    if stop_minikube or (
        not force and Confirm.ask("🛑 Do you want to stop Minikube now?")
    ):
        console.print(Panel("📴 Stopping Minikube", style="blue"))
        if not cleanup_manager.stop_minikube():
            console.print("❌ Failed to stop Minikube", style="red")
            sys.exit(1)
    else:
        console.print("🟢 Minikube left running", style="green")

    console.print("✅ Cleanup completed successfully!", style="green")


if __name__ == "__main__":
    main()
