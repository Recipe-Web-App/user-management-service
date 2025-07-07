#!/usr/bin/env python3
"""User Management Service Container Starter.

This script starts the user-management-service deployment in Kubernetes. It provides a
rich terminal interface with progress tracking.
"""

import sys
import time
from dataclasses import dataclass
from http import HTTPStatus
from typing import Any

import click
from kubernetes import client, config
from kubernetes.client.rest import ApiException
from rich.console import Console
from rich.panel import Panel
from rich.progress import Progress, SpinnerColumn, TaskID, TextColumn

from .minikube_utils import check_minikube_status

# Configure rich console
console = Console()


@dataclass
class ServiceConfig:
    """Configuration for the service starter."""

    namespace: str = "user-management"
    deployment_name: str = "user-management-deployment"


class ContainerStarter:
    """Manages starting the container deployment."""

    def __init__(self, config_obj: ServiceConfig) -> None:
        """Initialize the container starter."""
        self.config = config_obj
        self.api_client: client.ApiClient | None = None
        self.apps_v1_api: client.AppsV1Api | None = None

    def initialize_client(self) -> bool:
        """Initialize Kubernetes client."""
        try:
            config.load_kube_config()
            self.api_client = client.ApiClient()
            self.apps_v1_api = client.AppsV1Api(self.api_client)
        except Exception as e:
            console.print(
                f"❌ Failed to initialize Kubernetes client: {e}", style="red"
            )
            return False
        else:
            return True

    def check_deployment_exists(self) -> bool:
        """Check if deployment exists."""
        if self.apps_v1_api is None:
            return False
        try:
            self.apps_v1_api.read_namespaced_deployment(
                self.config.deployment_name, self.config.namespace
            )  # type: ignore[no-untyped-call]
        except ApiException as e:
            if e.status == HTTPStatus.NOT_FOUND:
                return False
            raise
        else:
            return True

    def scale_deployment(self, replicas: int) -> bool:
        """Scale deployment to specified number of replicas."""
        try:
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console,
            ) as progress:
                task = progress.add_task(
                    f"Scaling deployment to {replicas} replica(s)...", total=None
                )

                # Scale the deployment using Kubernetes client
                if self.apps_v1_api is None:
                    return False

                # Get current deployment
                deployment = self.apps_v1_api.read_namespaced_deployment(
                    name=self.config.deployment_name, namespace=self.config.namespace
                )  # type: ignore[no-untyped-call]

                # Update replicas
                deployment.spec.replicas = replicas
                self.apps_v1_api.patch_namespaced_deployment(
                    name=self.config.deployment_name,
                    namespace=self.config.namespace,
                    body=deployment,
                )  # type: ignore[no-untyped-call]

                progress.update(task, completed=True)
                return True
        except Exception as e:
            console.print(f"❌ Failed to scale deployment: {e}", style="red")
            return False

    def wait_for_deployment_ready(self, timeout: int = 90) -> bool:
        """Wait for deployment to be ready."""
        try:
            with Progress(
                SpinnerColumn(),
                TextColumn(text_format="[progress.description]{task.description}"),
                console=console,
            ) as progress:
                task: TaskID = progress.add_task(
                    description="Waiting for deployment to be ready...", total=None
                )

                # Wait for deployment to be ready using polling
                if self.apps_v1_api is None:
                    return False

                start_time = time.time()
                while time.time() - start_time < timeout:
                    try:
                        deployment = self.apps_v1_api.read_namespaced_deployment(
                            name=self.config.deployment_name,
                            namespace=self.config.namespace,
                        )  # type: ignore[no-untyped-call]

                        if (
                            deployment.status.available_replicas
                            == deployment.spec.replicas
                            and deployment.status.ready_replicas
                            == deployment.spec.replicas
                        ):
                            progress.update(task_id=task, completed=True)
                            return True

                        time.sleep(2)  # Poll every 2 seconds
                    except Exception:
                        time.sleep(2)
                        continue

                console.print(
                    "❌ Deployment failed to become ready within timeout", style="red"
                )
                return False
        except Exception as e:
            console.print(f"❌ Failed to wait for deployment: {e}", style="red")
            return False

    def get_deployment_status(self) -> dict[str, Any] | None:
        """Get current deployment status."""
        if self.apps_v1_api is None:
            return None
        try:
            deployment = self.apps_v1_api.read_namespaced_deployment(
                name=self.config.deployment_name, namespace=self.config.namespace
            )  # type: ignore[no-untyped-call]
        except ApiException as e:
            if e.status == HTTPStatus.NOT_FOUND:
                return None
            raise
        else:
            return {
                "name": deployment.metadata.name,
                "replicas": deployment.spec.replicas,
                "available": deployment.status.available_replicas,
                "ready": deployment.status.ready_replicas,
                "updated": deployment.status.updated_replicas,
            }


@click.command()
@click.option("--namespace", default="user-management", help="Kubernetes namespace")
@click.option(
    "--deployment", default="user-management-deployment", help="Deployment name"
)
@click.option(
    "--timeout", default=90, help="Timeout in seconds for deployment to be ready"
)
@click.option("--replicas", default=1, help="Number of replicas to scale to")
def main(namespace: str, deployment: str, timeout: int, replicas: int) -> None:
    """Start the user-management-service deployment."""
    # Initialize configuration
    config_obj = ServiceConfig(
        namespace=namespace,
        deployment_name=deployment,
    )

    # Check Minikube status
    console.print(Panel("🔍 Checking Minikube Status", style="blue"))
    if not check_minikube_status():
        console.print(
            "❌ Minikube is not running. Please start Minikube first.", style="red"
        )
        sys.exit(1)
    else:
        console.print("✅ Minikube is running", style="green")

    # Initialize container starter
    starter = ContainerStarter(config_obj)
    if not starter.initialize_client():
        console.print("❌ Failed to initialize Kubernetes client", style="red")
        sys.exit(1)

    # Check if deployment exists
    console.print(Panel("📦 Checking Deployment", style="blue"))
    if not starter.check_deployment_exists():
        console.print(
            f"❌ Deployment '{config_obj.deployment_name}' not found in namespace "
            f"'{config_obj.namespace}'",
            style="red",
        )
        console.print(
            "Please deploy the application first using the deploy script.",
            style="yellow",
        )
        sys.exit(1)
    else:
        console.print(
            f"✅ Deployment '{config_obj.deployment_name}' exists", style="green"
        )

    # Get current status
    current_status = starter.get_deployment_status()
    if current_status:
        console.print(f"Current replicas: {current_status['replicas']}", style="cyan")
        console.print(
            f"Available replicas: {current_status['available']}", style="cyan"
        )

    # Start the deployment
    console.print(Panel("🔄 Starting Container", style="blue"))
    if not starter.scale_deployment(replicas):
        console.print("❌ Failed to start deployment", style="red")
        sys.exit(1)

    # Wait for deployment to be ready
    console.print(Panel("⏳ Waiting for Deployment", style="blue"))
    if not starter.wait_for_deployment_ready(timeout):
        console.print("❌ Deployment failed to become ready", style="red")
        sys.exit(1)

    # Get final status
    final_status = starter.get_deployment_status()
    if final_status:
        console.print(Panel("📊 Final Deployment Status", style="green"))
        console.print(f"Name: {final_status['name']}", style="green")
        console.print(f"Replicas: {final_status['replicas']}", style="green")
        console.print(f"Available: {final_status['available']}", style="green")
        console.print(f"Ready: {final_status['ready']}", style="green")
        console.print(f"Updated: {final_status['updated']}", style="green")

    console.print("✅ Deployment started successfully!", style="green")


if __name__ == "__main__":
    main()
