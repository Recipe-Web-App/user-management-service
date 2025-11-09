"""Service-to-service communication constants.

This module contains hard-coded configuration for downstream service integrations. URLs
and scopes are defined here to avoid environment variable sprawl.
"""

from typing import Final

# Notification Service Configuration
NOTIFICATION_SERVICE_URL_LOCAL: Final[str] = (
    "http://notification-service.local/api/v1/notification"
)
NOTIFICATION_SERVICE_URL_K8S: Final[str] = (
    "http://notification-service.notification.svc.cluster.local:8000/api/v1/notification"
)
NOTIFICATION_SERVICE_SCOPES: Final[list[str]] = ["notification:admin"]


def get_notification_service_url(environment: str = "local") -> str:
    """Get the appropriate notification service URL based on environment.

    Args:
        environment: Deployment environment ("local" or "k8s")

    Returns:
        The notification service base URL for the specified environment

    Raises:
        ValueError: If environment is not "local" or "k8s"
    """
    if environment == "local":
        return NOTIFICATION_SERVICE_URL_LOCAL
    if environment == "k8s":
        return NOTIFICATION_SERVICE_URL_K8S
    raise ValueError(f"Invalid environment: {environment}. Must be 'local' or 'k8s'")
