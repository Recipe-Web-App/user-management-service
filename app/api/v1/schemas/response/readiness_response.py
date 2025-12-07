"""Readiness check response model."""

from pydantic import BaseModel, Field

from app.api.v1.schemas.response.dependency_health import DependencyHealth


class ReadinessResponse(BaseModel):
    """Response model for readiness checks."""

    ready: bool = Field(
        ..., description="Whether the service is ready to serve requests"
    )
    status: str = Field(..., description="Overall readiness status")
    degraded: bool = Field(
        default=False, description="Whether the service is running in degraded mode"
    )
    dependencies: dict[str, DependencyHealth] = Field(
        ..., description="Health status of each dependency"
    )
