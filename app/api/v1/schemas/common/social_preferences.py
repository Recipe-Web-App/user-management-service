"""Social preferences schema definitions."""

from pydantic import BaseModel, Field


class SocialPreferences(BaseModel):
    """Schema for user social preferences."""

    friend_requests: bool | None = Field(
        None, description="Whether friend requests are enabled"
    )
    message_notifications: bool | None = Field(
        None, description="Whether message notifications are enabled"
    )
    group_invites: bool | None = Field(
        None, description="Whether group invites are enabled"
    )
    share_activity: bool | None = Field(
        None, description="Whether sharing activity is enabled"
    )
