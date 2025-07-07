#!/usr/bin/env python3
"""User Management Service Container Deployment Script.

This script deploys the user-management-service to Kubernetes using Minikube. It
provides a rich terminal interface with progress tracking and status updates.
"""

import os
import sys
from dataclasses import dataclass
from http import HTTPStatus
from pathlib import Path

import click
import docker
import kubernetes
from dotenv import load_dotenv
from kubernetes import client, config
from kubernetes.client.rest import ApiException
from kubipy.utils import minipy
from rich.console import Console
from rich.layout import Layout
from rich.progress import Progress, SpinnerColumn, TextColumn

from .minikube_utils import check_minikube_status, start_minikube

# Configure rich console
console = Console()


@dataclass
class DeploymentConfig:
    """Configuration for the deployment."""

    namespace: str = "user-management"
    image_name: str = "user-management-service"
    image_tag: str = "latest"
    config_dir: str = "k8s"
    secret_name: str = "user-management-db-password"
    deployment_name: str = "user-management-deployment"
    service_name: str = "user-management-service"

    @property
    def full_image_name(self) -> str:
        """Get the full Docker image name with tag."""
        return f"{self.image_name}:{self.image_tag}"


class KubernetesManager:
    """Manages Kubernetes operations."""

    def __init__(self, config_obj: DeploymentConfig) -> None:
        """Initialize the Kubernetes manager."""
        self.config = config_obj
        self.api_client: client.ApiClient | None = None
        self.core_v1_api: client.CoreV1Api | None = None
        self.apps_v1_api: client.AppsV1Api | None = None

    def initialize_client(self) -> bool:
        """Initialize Kubernetes client."""
        try:
            # Try to load from kubeconfig first
            config.load_kube_config()
            console.print("✅ Loaded Kubernetes config from kubeconfig", style="green")
        except Exception:
            try:
                # Try to load from service account
                config.load_incluster_config()
                console.print(
                    "✅ Loaded Kubernetes config from service account", style="green"
                )
            except Exception as e:
                console.print(f"❌ Failed to load Kubernetes config: {e}", style="red")
                return False
            else:
                self.api_client = client.ApiClient()
                self.core_v1_api = client.CoreV1Api(self.api_client)
                self.apps_v1_api = client.AppsV1Api(self.api_client)
                return True
        else:
            self.api_client = client.ApiClient()
            self.core_v1_api = client.CoreV1Api(self.api_client)
            self.apps_v1_api = client.AppsV1Api(self.api_client)
            return True

    def check_namespace_exists(self) -> bool:
        """Check if namespace exists."""
        if self.core_v1_api is None:
            return False
        try:
            self.core_v1_api.read_namespace(
                self.config.namespace
            )  # type: ignore[no-untyped-call]
        except ApiException as e:
            if e.status == HTTPStatus.NOT_FOUND:
                return False
            raise
        else:
            return True

    def create_namespace(self) -> bool:
        """Create namespace if it doesn't exist."""
        if self.check_namespace_exists():
            console.print(
                f"✅ Namespace '{self.config.namespace}' already exists", style="green"
            )
            return True

        if self.core_v1_api is None:
            return False
        try:
            namespace = client.V1Namespace(
                metadata=client.V1ObjectMeta(name=self.config.namespace)
            )
            self.core_v1_api.create_namespace(
                namespace
            )  # type: ignore[no-untyped-call]
            console.print(
                f"✅ Created namespace '{self.config.namespace}'", style="green"
            )
        except Exception as e:
            console.print(f"❌ Failed to create namespace: {e}", style="red")
            return False
        else:
            return True


class EnvironmentManager:
    """Manages environment variables and configuration."""

    def __init__(self) -> None:
        """Initialize the environment manager."""
        self.env_vars: dict[str, str] = {}

    def load_env_file(self, env_file: str = ".env") -> bool:
        """Load environment variables from .env file."""
        env_path = Path(env_file)
        if not env_path.exists():
            console.print(f"⚠️ No .env file found at {env_file}", style="yellow")
            return False

        try:
            load_dotenv(env_path)
            # Get all environment variables
            self.env_vars = dict(os.environ)
            console.print(
                f"✅ Loaded {len(self.env_vars)} environment variables from {env_file}",
                style="green",
            )
        except Exception as e:
            console.print(f"❌ Failed to load .env file: {e}", style="red")
            return False
        else:
            return True

    def get_required_vars(self) -> dict[str, str]:
        """Get required environment variables."""
        required_vars = [
            "POSTGRES_HOST",
            "POSTGRES_PORT",
            "POSTGRES_DB",
            "POSTGRES_SCHEMA",
            "USER_MANAGEMENT_DB_USER",
            "USER_MANAGEMENT_DB_PASSWORD",
            "JWT_SECRET_KEY",
            "JWT_SIGNING_ALGORITHM",
            "ACCESS_TOKEN_EXPIRE_MINUTES",
            "ALLOWED_ORIGIN_HOSTS",
            "ALLOWED_CREDENTIALS",
        ]

        missing_vars: list[str] = []
        found_vars: dict[str, str] = {}

        for var in required_vars:
            if var in self.env_vars:
                found_vars[var] = self.env_vars[var]
            else:
                missing_vars.append(var)

        if missing_vars:
            console.print(
                f"⚠️ Missing environment variables: {', '.join(missing_vars)}",
                style="yellow",
            )

        return found_vars


class DockerManager:
    """Manages Docker operations."""

    def __init__(self, config_obj: DeploymentConfig) -> None:
        """Initialize the Docker manager."""
        self.config = config_obj

    def build_image(self) -> bool:
        """Build Docker image."""
        try:
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console,
            ) as progress:
                task = progress.add_task("Building Docker image...", total=None)

                # Use kubipy to set minikube docker environment
                mk = minipy()
                mk.docker_env()

                # Execute docker build using docker-py
                client = docker.from_env()
                client.images.build(path=".", tag=self.config.full_image_name, rm=True)

                progress.update(task, completed=True)
                console.print("✅ Docker image built successfully", style="green")
                return True
        except Exception as e:
            console.print(f"❌ Failed to build Docker image: {e}", style="red")
            return False


def check_prerequisites() -> bool:
    """Check if all prerequisites are met."""
    # Check if minikube is available via kubipy
    try:
        mk = minipy()
        mk.status()
    except Exception:
        console.print("❌ Minikube is not available", style="red")
        return False

    # Check if kubernetes client is available
    try:
        kubernetes.config.load_kube_config()
    except Exception:
        console.print("❌ Kubernetes client is not available", style="red")
        return False

    # Check if docker is available
    try:
        docker.from_env()
    except Exception:
        console.print("❌ Docker is not available", style="red")
        return False

    console.print("✅ All prerequisites are met", style="green")
    return True


def create_deployment_layout() -> Layout:
    """Create the deployment status layout."""
    layout = Layout()
    layout.split_column(
        Layout(name="header", size=3),
        Layout(name="main"),
        Layout(name="footer", size=3),
    )
    return layout


def load_environment(env_file: str) -> dict[str, str]:
    """Load environment variables from the given .env file."""
    env_manager = EnvironmentManager()
    env_manager.load_env_file(env_file)
    return env_manager.env_vars


def start_minikube_or_exit() -> None:
    """Start Minikube or exit if it fails."""
    if check_minikube_status():
        console.print("✅ Minikube is already running", style="green")
        return

    console.print("🔄 Starting Minikube...", style="blue")
    if not start_minikube():
        console.print("❌ Failed to start Minikube", style="red")
        sys.exit(1)
    console.print("✅ Minikube started", style="green")


def initialize_k8s_or_exit(config_obj: DeploymentConfig) -> KubernetesManager:
    """Initialize the Kubernetes client or exit if it fails."""
    k8s_manager = KubernetesManager(config_obj)
    if not k8s_manager.initialize_client():
        console.print("❌ Failed to initialize Kubernetes client", style="red")
        sys.exit(1)
    return k8s_manager


def create_namespace_or_exit(k8s_manager: KubernetesManager) -> None:
    """Create the Kubernetes namespace or exit if it fails."""
    if not k8s_manager.create_namespace():
        console.print("❌ Failed to create namespace", style="red")
        sys.exit(1)


def build_docker_image_or_exit(config_obj: DeploymentConfig) -> None:
    """Build the Docker image or exit if it fails."""
    docker_manager = DockerManager(config_obj)
    if not docker_manager.build_image():
        console.print("❌ Failed to build Docker image", style="red")
        sys.exit(1)


def apply_manifests_or_exit(config_obj: DeploymentConfig, namespace: str) -> None:
    """Apply Kubernetes manifests or exit if any fail."""
    k8s_dir = Path(config_obj.config_dir)
    if not k8s_dir.exists():
        console.print(
            f"❌ Kubernetes config directory not found: {k8s_dir}", style="red"
        )
        sys.exit(1)
    manifests = [
        "configmap-template.yaml",
        "secret-template.yaml",
        "deployment.yaml",
        "service.yaml",
        "ingress.yaml",
    ]
    for manifest in manifests:
        manifest_path = k8s_dir / manifest
        if manifest_path.exists():
            try:
                # Use kubipy to apply manifests
                mk = minipy()
                mk.kubectl(["apply", "-f", str(manifest_path), "-n", namespace])
                console.print(f"✅ Applied {manifest}", style="green")
            except Exception as e:
                console.print(f"❌ Failed to apply {manifest}: {e}", style="red")
                sys.exit(1)
        else:
            console.print(f"⚠️ Manifest not found: {manifest}", style="yellow")


def wait_for_deployment_ready_or_exit(
    config_obj: DeploymentConfig, namespace: str
) -> None:
    """Wait for the deployment to be ready or exit if it fails."""
    console.print("⏳ Waiting for deployment to be ready...", style="blue")
    try:
        # Use kubipy to wait for deployment
        mk = minipy()
        mk.kubectl(
            [
                "wait",
                "--for=condition=available",
                f"deployment/{config_obj.deployment_name}",
                "-n",
                namespace,
                "--timeout=300s",
            ]
        )
        console.print("✅ Deployment is ready", style="green")
    except Exception:
        console.print("❌ Deployment failed to become ready", style="red")
        sys.exit(1)


def print_service_url(config_obj: DeploymentConfig, namespace: str) -> None:
    """Print the service URL for the deployed service."""
    try:
        # Use kubipy to get service URL
        mk = minipy()
        service_url = mk.service(config_obj.service_name, namespace=namespace, url=True)
        console.print(f"🌐 Service URL: {service_url}", style="green")
    except Exception:
        console.print("⚠️ Could not get service URL", style="yellow")


@click.command()
@click.option("--env-file", default=".env", help="Path to environment file")
@click.option("--namespace", default="user-management", help="Kubernetes namespace")
@click.option("--image-tag", default="latest", help="Docker image tag")
@click.option(
    "--dry-run",
    is_flag=True,
    help="Show what would be deployed without actually deploying",
)
def main(env_file: str, namespace: str, image_tag: str, dry_run: bool) -> None:
    """Deploy the user-management-service to Kubernetes."""
    config_obj = DeploymentConfig(
        namespace=namespace,
        image_tag=image_tag,
    )
    if not check_prerequisites():
        sys.exit(1)
    load_environment(env_file)
    start_minikube_or_exit()
    k8s_manager = initialize_k8s_or_exit(config_obj)
    create_namespace_or_exit(k8s_manager)
    if dry_run:
        console.print("🔍 DRY RUN MODE - No changes will be made", style="yellow")
        console.print("Would deploy with the following configuration:")
        console.print(f"  Namespace: {config_obj.namespace}")
        console.print(f"  Image: {config_obj.full_image_name}")
        console.print(f"  Config directory: {config_obj.config_dir}")
        return
    build_docker_image_or_exit(config_obj)
    console.print("🚀 Deploying to Kubernetes...", style="blue")
    apply_manifests_or_exit(config_obj, namespace)
    wait_for_deployment_ready_or_exit(config_obj, namespace)
    print_service_url(config_obj, namespace)
    console.print("🎉 Deployment completed successfully!", style="green")


if __name__ == "__main__":
    main()
