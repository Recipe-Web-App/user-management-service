"""Sound preferences schema definitions."""

from pydantic import BaseModel, Field


class SoundPreferences(BaseModel):
    """Schema for user sound preferences."""

    notification_sounds: bool = Field(
        ..., description="Whether notification sounds are enabled"
    )
    system_sounds: bool = Field(..., description="Whether system sounds are enabled")
    volume_level: bool = Field(..., description="Whether volume level is enabled")
    mute_notifications: bool = Field(
        ..., description="Whether notification muting is enabled"
    )
