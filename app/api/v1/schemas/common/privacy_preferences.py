"""Privacy preferences schema definitions."""

from pydantic import BaseModel, Field

from app.enums.preferences.profile_visibility_enum import ProfileVisibilityEnum


class PrivacyPreferences(BaseModel):
    """Schema for user privacy preferences."""

    profile_visibility: ProfileVisibilityEnum = Field(
        ..., description="Profile visibility setting"
    )
    recipe_visibility: ProfileVisibilityEnum = Field(
        ..., description="Recipe visibility setting"
    )
    activity_visibility: ProfileVisibilityEnum = Field(
        ..., description="Activity visibility setting"
    )
    contact_info_visibility: ProfileVisibilityEnum = Field(
        ..., description="Contact info visibility setting"
    )
    data_sharing: bool = Field(..., description="Whether data sharing is enabled")
    analytics_tracking: bool = Field(
        ..., description="Whether analytics tracking is enabled"
    )
