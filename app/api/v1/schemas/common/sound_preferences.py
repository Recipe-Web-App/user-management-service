"""Sound preferences schema definitions."""

from pydantic import BaseModel, Field


class SoundPreferences(BaseModel):
    """Schema for user sound preferences."""

    notification_sounds: bool | None = Field(
        None, description="Whether notification sounds are enabled"
    )
    system_sounds: bool | None = Field(
        None, description="Whether system sounds are enabled"
    )
    volume_level: bool | None = Field(
        None, description="Whether volume level is enabled"
    )
    mute_notifications: bool | None = Field(
        None, description="Whether notification muting is enabled"
    )
