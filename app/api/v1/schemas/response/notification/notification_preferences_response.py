"""Notification preferences response schema."""

from typing import Any

from pydantic import BaseModel, Field


class NotificationPreferencesResponse(BaseModel):
    """Response schema for notification preferences."""

    user_id: str = Field(..., description="User ID")
    preferences: dict[str, Any] = Field(
        ..., description="Notification preferences key-value pairs"
    )
    message: str = Field(
        default="Notification preferences retrieved successfully",
        description="Response message",
    )
