#!/usr/bin/env python3
"""Minikube utility functions for container management scripts."""

from functools import lru_cache
from typing import Any

from kubipy.utils import minipy


def check_minikube_status() -> bool:
    """Check if Minikube is running.

    Returns:
        bool: True if Minikube is running, False otherwise.
    """
    try:
        mk = get_minikube_instance()
        status = mk.current_status()
        return bool(status and "Running" in status)
    except Exception:
        return False


def get_minikube_status_info() -> dict[str, str]:
    """Get detailed Minikube status information.

    Returns:
        dict[str, str]: Dictionary containing parsed status information.
    """
    try:
        mk = get_minikube_instance()
        status = mk.current_status()
        # Parse the status output
        status_dict: dict[str, str] = {}
        if status:
            # The status method returns a string, parse it
            for line in status.split("\n"):
                if ":" in line:
                    key, value = line.split(":", 1)
                    status_dict[key.strip()] = value.strip()
            return status_dict
    except Exception:
        return {"Status": "Not Running"}
    else:
        return {"Status": "Not Running"}


def start_minikube() -> bool:
    """Start Minikube if not running.

    Returns:
        bool: True if Minikube was started successfully or already running.
    """
    if check_minikube_status():
        return True

    try:
        mk = get_minikube_instance()
        mk.start()
    except Exception:
        return False
    else:
        return True


def stop_minikube() -> bool:
    """Stop Minikube.

    Returns:
        bool: True if Minikube was stopped successfully.
    """
    try:
        mk = get_minikube_instance()
        mk.stop()
    except Exception:
        return False
    else:
        return True


@lru_cache(maxsize=1)
def get_minikube_instance() -> Any:
    """Get a Minikube instance.

    Returns:
        The Minikube instance from kubipy.
    """
    return minipy()
