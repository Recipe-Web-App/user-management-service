"""Privacy preferences schema definitions."""

from pydantic import BaseModel, Field

from app.enums.preferences.profile_visibility_enum import ProfileVisibilityEnum


class PrivacyPreferences(BaseModel):
    """Schema for user privacy preferences."""

    profile_visibility: ProfileVisibilityEnum | None = Field(
        None, description="Profile visibility setting"
    )
    recipe_visibility: ProfileVisibilityEnum | None = Field(
        None, description="Recipe visibility setting"
    )
    activity_visibility: ProfileVisibilityEnum | None = Field(
        None, description="Activity visibility setting"
    )
    contact_info_visibility: ProfileVisibilityEnum | None = Field(
        None, description="Contact info visibility setting"
    )
    data_sharing: bool | None = Field(
        None, description="Whether data sharing is enabled"
    )
    analytics_tracking: bool | None = Field(
        None, description="Whether analytics tracking is enabled"
    )
