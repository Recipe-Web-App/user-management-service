"""Redis session statistics response schema.

Provides the response model for Redis session statistics endpoint.
"""

from pydantic import BaseModel, Field


class RedisSessionStatsResponse(BaseModel):
    """Response schema for Redis session statistics."""

    total_sessions: int = Field(..., description="Total number of sessions in Redis.")
    active_sessions: int = Field(..., description="Number of active sessions.")
    expired_sessions: int = Field(..., description="Number of expired sessions.")
    memory_usage_bytes: int = Field(..., description="Redis used memory in bytes.")
    memory_usage_mb: float = Field(..., description="Redis used memory in MB.")
    key_count: int = Field(..., description="Total number of keys in Redis DB.")
    session_key_count: int = Field(..., description="Number of session keys in Redis.")
    session_ttl_min: int | None = Field(
        None, description="Minimum TTL (seconds) among session keys."
    )
    session_ttl_max: int | None = Field(
        None, description="Maximum TTL (seconds) among session keys."
    )
    session_ttl_avg: float | None = Field(
        None, description="Average TTL (seconds) among session keys."
    )
    redis_uptime_seconds: int = Field(
        ..., description="Redis server uptime in seconds."
    )
    redis_version: str = Field(..., description="Redis server version.")
