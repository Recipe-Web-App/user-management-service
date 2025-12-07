"""Force logout response schema.

Provides the response model for force logout endpoint.
"""

from datetime import datetime
from uuid import UUID

from pydantic import BaseModel, Field


class ForceLogoutResponse(BaseModel):
    """Response schema for force logout operation."""

    user_id: UUID = Field(..., description="User ID that was force logged out.")
    sessions_terminated: int = Field(..., description="Number of sessions terminated.")
    timestamp: datetime = Field(
        ..., description="Timestamp when force logout was executed."
    )
    admin_user_id: UUID = Field(
        ..., description="Admin user ID who executed the force logout."
    )
