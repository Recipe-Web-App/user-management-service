"""Theme preferences schema definitions."""

from pydantic import BaseModel, Field

from app.enums.preferences.theme_enum import ThemeEnum


class ThemePreferences(BaseModel):
    """Schema for user theme preferences."""

    dark_mode: bool | None = Field(None, description="Whether dark mode is enabled")
    light_mode: bool | None = Field(None, description="Whether light mode is enabled")
    auto_theme: bool | None = Field(None, description="Whether auto theme is enabled")
    custom_theme: ThemeEnum | None = Field(
        None, description="Custom theme name, if set"
    )
