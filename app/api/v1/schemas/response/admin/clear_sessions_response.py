"""Clear sessions response schema.

Provides the response model for clear sessions endpoint.
"""

from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, Field


class ClearSessionsResponse(BaseModel):
    """Response schema for clear sessions operation."""

    sessions_cleared: int = Field(..., description="Number of sessions cleared.")
    timestamp: datetime = Field(
        ..., description="Timestamp when sessions were cleared."
    )
    admin_user_id: UUID = Field(
        ..., description="Admin user ID who executed the clear sessions."
    )
    redis_keys_removed: int = Field(..., description="Number of Redis keys removed.")
