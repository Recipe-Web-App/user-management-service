"""System health response schema.

Provides the response model for system health endpoint.
"""

from datetime import datetime

from pydantic import BaseModel, Field

from app.enums.environment_enum import EnvironmentEnum


class SystemHealthResponse(BaseModel):
    """Response schema for system health status."""

    database: bool = Field(..., description="Database connection status.")
    redis: bool = Field(..., description="Redis connection status.")
    uptime_seconds: int = Field(..., description="Application uptime in seconds.")
    version: str = Field(..., description="Application version.")
    environment: EnvironmentEnum = Field(..., description="Current environment.")
    timestamp: datetime = Field(..., description="Health check timestamp.")
