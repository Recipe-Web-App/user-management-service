"""Social preferences schema definitions."""

from pydantic import BaseModel, Field


class SocialPreferences(BaseModel):
    """Schema for user social preferences."""

    friend_requests: bool = Field(
        ..., description="Whether friend requests are enabled"
    )
    message_notifications: bool = Field(
        ..., description="Whether message notifications are enabled"
    )
    group_invites: bool = Field(..., description="Whether group invites are enabled")
    share_activity: bool = Field(..., description="Whether sharing activity is enabled")
