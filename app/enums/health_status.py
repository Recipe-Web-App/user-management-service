"""Health status enumeration for dependency checks."""

from enum import Enum


class HealthStatus(str, Enum):
    """Health status enumeration for service dependencies."""

    HEALTHY = "healthy"
    DEGRADED = "degraded"
    UNHEALTHY = "unhealthy"
    TIMEOUT = "timeout"
    DISCONNECTED = "disconnected"
    ERROR = "error"
