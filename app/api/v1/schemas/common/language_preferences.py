"""Language preferences schema definitions."""

from pydantic import BaseModel, Field

from app.enums.preferences.language_enum import LanguageEnum


class LanguagePreferences(BaseModel):
    """Schema for user language preferences."""

    primary_language: LanguageEnum | None = Field(
        None, description="Primary language preference"
    )
    secondary_language: LanguageEnum | None = Field(
        None, description="Secondary language preference, if set"
    )
    translation_enabled: bool | None = Field(
        None, description="Whether translation is enabled"
    )
