"""Dependency health response model."""

from pydantic import BaseModel, Field

from app.enums.health_status import HealthStatus


class DependencyHealth(BaseModel):
    """Health status for a single dependency."""

    healthy: bool = Field(..., description="Whether the dependency is healthy")
    status: HealthStatus = Field(..., description="Detailed status of the dependency")
    message: str = Field(..., description="Human-readable status message")
    response_time_ms: float | None = Field(
        None, description="Response time in milliseconds"
    )
