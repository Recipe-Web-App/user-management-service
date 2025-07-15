"""Accessibility preferences schema definitions."""

from pydantic import BaseModel, Field


class AccessibilityPreferences(BaseModel):
    """Schema for user accessibility preferences."""

    screen_reader: bool = Field(
        ..., description="Whether screen reader support is enabled"
    )
    high_contrast: bool = Field(
        ..., description="Whether high contrast mode is enabled"
    )
    reduced_motion: bool = Field(..., description="Whether reduced motion is enabled")
    large_text: bool = Field(..., description="Whether large text is enabled")
    keyboard_navigation: bool = Field(
        ..., description="Whether keyboard navigation is enabled"
    )
