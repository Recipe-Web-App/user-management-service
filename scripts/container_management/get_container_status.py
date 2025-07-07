#!/usr/bin/env python3
"""User Management Service Container Status Checker.

This script checks the status of the user-management-service deployment in Kubernetes.
It provides a rich terminal interface with detailed status information.
"""

import sys
from dataclasses import dataclass
from http import HTTPStatus
from typing import Any

import click
from kubernetes import client, config
from kubernetes.client.rest import ApiException
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

from .minikube_utils import get_minikube_status_info

# Configure rich console
console = Console()


@dataclass
class ServiceConfig:
    """Configuration for the service status checker."""

    namespace: str = "user-management"
    deployment_name: str = "user-management-deployment"
    service_name: str = "user-management-service"
    ingress_name: str = "user-management-ingress"


class StatusChecker:
    """Checks the status of Kubernetes resources."""

    def __init__(self, config_obj: ServiceConfig) -> None:
        """Initialize the status checker."""
        self.config = config_obj
        self.api_client: client.ApiClient
        self.core_v1_api: client.CoreV1Api
        self.apps_v1_api: client.AppsV1Api
        self.networking_v1_api: client.NetworkingV1Api

    def initialize_client(self) -> bool:
        """Initialize Kubernetes client."""
        try:
            config.load_kube_config()
            self.api_client = client.ApiClient()
            self.core_v1_api = client.CoreV1Api(self.api_client)
            self.apps_v1_api = client.AppsV1Api(self.api_client)
            self.networking_v1_api = client.NetworkingV1Api(self.api_client)
        except Exception as e:
            console.print(
                f"❌ Failed to initialize Kubernetes client: {e}", style="red"
            )
            return False
        else:
            return True

    def check_namespace(self) -> bool:
        """Check if namespace exists."""
        if self.core_v1_api is None:
            return False
        try:
            self.core_v1_api.read_namespace(self.config.namespace)  # type: ignore
        except ApiException as e:
            if e.status == HTTPStatus.NOT_FOUND:
                return False
            raise
        else:
            return True

    def check_deployment(self) -> dict[str, Any] | None:
        """Check deployment status."""
        if self.apps_v1_api is None:
            return None
        try:
            deployment = self.apps_v1_api.read_namespaced_deployment(
                self.config.deployment_name, self.config.namespace
            )  # type: ignore
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

    def check_service(self) -> dict[str, Any] | None:
        """Check service status."""
        if self.core_v1_api is None:
            return None
        try:
            service = self.core_v1_api.read_namespaced_service(
                self.config.service_name, self.config.namespace
            )  # type: ignore
            return {
                "name": service.metadata.name,
                "type": service.spec.type,
                "cluster_ip": service.spec.cluster_ip,
                "ports": (
                    [port.port for port in service.spec.ports]
                    if service.spec.ports
                    else []
                ),
            }
        except ApiException as e:
            if e.status == HTTPStatus.NOT_FOUND:
                return None
            raise

    def check_ingress(self) -> dict[str, Any] | None:
        """Check ingress status."""
        # TODO: Fix ingress check when proper method is available
        return None

    def check_pods(self) -> list[dict[str, Any]]:
        """Check pod status."""
        if self.core_v1_api is None:
            return []
        try:
            pods = self.core_v1_api.list_namespaced_pod(
                self.config.namespace,
                label_selector=f"app={self.config.deployment_name}",
            )  # type: ignore
            pod_statuses = []
            for pod in pods.items:
                status = {
                    "name": pod.metadata.name,
                    "phase": pod.status.phase,
                    "ready": pod.status.ready,
                    "restarts": (
                        pod.status.container_statuses[0].restart_count
                        if pod.status.container_statuses
                        else 0
                    ),
                }
                pod_statuses.append(status)
        except Exception as e:
            console.print(f"❌ Failed to check pods: {e}", style="red")
            return []
        else:
            return pod_statuses

    def check_persistent_volumes(self) -> list[dict[str, Any]]:
        """Check PVC status."""
        if self.core_v1_api is None:
            return []
        try:
            pvcs = self.core_v1_api.list_namespaced_persistent_volume_claim(
                self.config.namespace
            )  # type: ignore
            pvc_statuses = []
            for pvc in pvcs.items:
                status = {
                    "name": pvc.metadata.name,
                    "status": pvc.status.phase,
                    "capacity": (
                        pvc.status.capacity.get("storage", "Unknown")
                        if pvc.status.capacity
                        else "Unknown"
                    ),
                }
                pvc_statuses.append(status)
        except Exception as e:
            console.print(f"❌ Failed to check PVCs: {e}", style="red")
            return []
        else:
            return pvc_statuses


def display_minikube_status() -> None:
    """Display Minikube status."""
    console.print(Panel("🔍 Minikube Status", style="blue"))

    status = get_minikube_status_info()
    status_table = Table(show_header=False)
    status_table.add_column("Property", style="cyan")
    status_table.add_column("Value", style="green")

    for key, value in status.items():
        status_table.add_row(key, value)

    console.print(status_table)


def display_namespace_status(status_checker: StatusChecker) -> None:
    """Display namespace status."""
    console.print(Panel("📁 Namespace Status", style="blue"))

    if status_checker.check_namespace():
        console.print(
            f"✅ Namespace '{status_checker.config.namespace}' exists", style="green"
        )
    else:
        console.print(
            f"❌ Namespace '{status_checker.config.namespace}' does not exist",
            style="red",
        )


def display_deployment_status(status_checker: StatusChecker) -> None:
    """Display deployment status."""
    console.print(Panel("📦 Deployment Status", style="blue"))

    deployment = status_checker.check_deployment()
    if deployment:
        deployment_table = Table(title="Deployment Information")
        deployment_table.add_column("Property", style="cyan")
        deployment_table.add_column("Value", style="green")

        deployment_table.add_row("Name", deployment["name"])
        deployment_table.add_row("Replicas", str(deployment["replicas"]))
        deployment_table.add_row("Available", str(deployment["available"]))
        deployment_table.add_row("Ready", str(deployment["ready"]))
        deployment_table.add_row("Updated", str(deployment["updated"]))

        console.print(deployment_table)
    else:
        console.print(
            f"❌ Deployment '{status_checker.config.deployment_name}' not found",
            style="red",
        )


def display_service_status(status_checker: StatusChecker) -> None:
    """Display service status."""
    console.print(Panel("🌐 Service Status", style="blue"))

    service = status_checker.check_service()
    if service:
        service_table = Table(title="Service Information")
        service_table.add_column("Property", style="cyan")
        service_table.add_column("Value", style="green")

        service_table.add_row("Name", service["name"])
        service_table.add_row("Type", service["type"])
        service_table.add_row("Cluster IP", service["cluster_ip"])
        service_table.add_row("Ports", ", ".join(map(str, service["ports"])))

        console.print(service_table)
    else:
        console.print(
            f"❌ Service '{status_checker.config.service_name}' not found", style="red"
        )


def display_ingress_status(status_checker: StatusChecker) -> None:
    """Display ingress status."""
    console.print(Panel("🔗 Ingress Status", style="blue"))

    ingress = status_checker.check_ingress()
    if ingress:
        ingress_table = Table(title="Ingress Information")
        ingress_table.add_column("Property", style="cyan")
        ingress_table.add_column("Value", style="green")

        ingress_table.add_row("Name", ingress["name"])
        ingress_table.add_row("Class", ingress["class"])
        ingress_table.add_row("Hosts", ", ".join(ingress["hosts"]))

        console.print(ingress_table)
    else:
        console.print(
            f"❌ Ingress '{status_checker.config.ingress_name}' not found", style="red"
        )


def display_pod_status(status_checker: StatusChecker) -> None:
    """Display pod status."""
    console.print(Panel("🐳 Pod Status", style="blue"))

    pods = status_checker.check_pods()
    if pods:
        pod_table = Table(title="Pod Information")
        pod_table.add_column("Name", style="cyan")
        pod_table.add_column("Phase", style="green")
        pod_table.add_column("Ready", style="yellow")
        pod_table.add_column("Restarts", style="red")

        for pod in pods:
            pod_table.add_row(
                pod["name"], pod["phase"], str(pod["ready"]), str(pod["restarts"])
            )

        console.print(pod_table)
    else:
        console.print("❌ No pods found", style="red")


def display_pvc_status(status_checker: StatusChecker) -> None:
    """Display PVC status."""
    console.print(Panel("💾 Persistent Volume Claims", style="blue"))

    pvcs = status_checker.check_persistent_volumes()
    if pvcs:
        pvc_table = Table(title="PVC Information")
        pvc_table.add_column("Name", style="cyan")
        pvc_table.add_column("Status", style="green")
        pvc_table.add_column("Capacity", style="yellow")

        for pvc in pvcs:
            pvc_table.add_row(pvc["name"], pvc["status"], pvc["capacity"])

        console.print(pvc_table)
    else:
        console.print("❌ No PVCs found", style="red")


@click.command()
@click.option("--namespace", default="user-management", help="Kubernetes namespace")
def main(namespace: str) -> None:
    """Check the status of the user-management-service deployment."""
    # Initialize configuration
    config_obj = ServiceConfig(namespace=namespace)

    # Initialize status checker
    status_checker = StatusChecker(config_obj)
    if not status_checker.initialize_client():
        sys.exit(1)

    # Display status information
    display_minikube_status()
    console.print()

    display_namespace_status(status_checker)
    console.print()

    display_deployment_status(status_checker)
    console.print()

    display_service_status(status_checker)
    console.print()

    display_ingress_status(status_checker)
    console.print()

    display_pod_status(status_checker)
    console.print()

    display_pvc_status(status_checker)
    console.print()

    console.print("📊 Container status check complete.", style="green")


if __name__ == "__main__":
    main()
