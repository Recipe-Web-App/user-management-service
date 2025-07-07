#!/usr/bin/env python3
"""Minikube utility functions for container management scripts."""

import os
import shutil
import subprocess  # nosec B404

from rich.console import Console

console = Console()


def get_executable_path(name: str) -> str:
    """Get full path to executable with validation.

    Args:
        name: Name of the executable to find

    Returns:
        Full path to the executable

    Raises:
        RuntimeError: If executable is not found or not executable
    """
    path = shutil.which(name)
    if not path:
        raise RuntimeError("Executable '" + name + "' not found in PATH")

    # Additional validation
    if not os.access(path, os.X_OK):
        raise RuntimeError("Executable '" + path + "' is not executable")

    return path


def check_minikube_status() -> bool:
    """Check if Minikube is running.

    Returns:
        bool: True if Minikube is running, False otherwise.
    """
    try:
        minikube_path = get_executable_path("minikube")
        result = subprocess.run(  # nosec B603
            [minikube_path, "status"],
            capture_output=True,
            text=True,
            check=False,
        )

        # Check if the command succeeded
        if result.returncode != 0:
            return False

        # Parse the status output to check all components
        status_lines = result.stdout.strip().split("\n")
        required_components = ["host", "kubelet", "apiserver", "kubeconfig"]
        running_components = 0

        for line in status_lines:
            if ":" in line:
                component, status = line.split(":", 1)
                component = component.strip().lower()
                status = status.strip().lower()

                if component in required_components and status == "running":
                    running_components += 1

        # All required components must be running
        return running_components == len(required_components)
    except Exception:
        return False


def get_minikube_status_info() -> dict[str, str]:
    """Get detailed Minikube status information.

    Returns:
        dict[str, str]: Dictionary containing parsed status information.
    """
    try:
        minikube_path = get_executable_path("minikube")
        result = subprocess.run(  # nosec B603
            [minikube_path, "status"],
            capture_output=True,
            text=True,
            check=False,
        )

        status_dict: dict[str, str] = {}
        if result.returncode == 0:
            for line in result.stdout.split("\n"):
                if ":" in line:
                    key, value = line.split(":", 1)
                    status_dict[key.strip()] = value.strip()
            return status_dict
    except (OSError, subprocess.SubprocessError):
        pass

    return {"Status": "Not Running"}


def start_minikube() -> bool:
    """Start Minikube if not running.

    Returns:
        bool: True if Minikube was started successfully or already running.
    """
    if check_minikube_status():
        return True

    try:
        # Start minikube with wait flag to ensure it's fully ready
        console.print("🚀 Starting Minikube...", style="blue")
        minikube_path = get_executable_path("minikube")
        subprocess.run(  # nosec B603
            [minikube_path, "start", "--wait=true", "--wait-timeout=300s"],
            check=True,
        )
        console.print("✅ Minikube started successfully", style="green")
    except subprocess.CalledProcessError as e:
        console.print(f"❌ Failed to start Minikube: {e}", style="red")
        return False
    except Exception as e:
        console.print(f"❌ Unexpected error starting Minikube: {e}", style="red")
        return False
    else:
        return True


def stop_minikube() -> bool:
    """Stop Minikube.

    Returns:
        bool: True if Minikube was stopped successfully.
    """
    try:
        minikube_path = get_executable_path("minikube")
        subprocess.run([minikube_path, "stop"], check=True)  # nosec B603
    except subprocess.CalledProcessError as e:
        console.print(f"❌ Failed to stop Minikube: {e}", style="red")
        return False
    except Exception as e:
        console.print(f"❌ Unexpected error stopping Minikube: {e}", style="red")
        return False
    else:
        return True
