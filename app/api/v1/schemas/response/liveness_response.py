"""Liveness check response model."""

from pydantic import BaseModel, Field


class LivenessResponse(BaseModel):
    """Response model for liveness checks."""

    alive: bool = Field(..., description="Whether the service is alive")
    status: str = Field(..., description="Liveness status")
    message: str = Field(..., description="Status message")
