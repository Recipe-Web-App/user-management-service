"""Accessibility preferences schema definitions."""

from pydantic import BaseModel, Field


class AccessibilityPreferences(BaseModel):
    """Schema for user accessibility preferences."""

    screen_reader: bool | None = Field(
        None, description="Whether screen reader support is enabled"
    )
    high_contrast: bool | None = Field(
        None, description="Whether high contrast mode is enabled"
    )
    reduced_motion: bool | None = Field(
        None, description="Whether reduced motion is enabled"
    )
    large_text: bool | None = Field(None, description="Whether large text is enabled")
    keyboard_navigation: bool | None = Field(
        None, description="Whether keyboard navigation is enabled"
    )
