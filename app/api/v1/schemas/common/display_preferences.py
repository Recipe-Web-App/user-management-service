"""Display preferences schema definitions."""

from pydantic import BaseModel, Field

from app.enums.preferences.color_scheme_enum import ColorSchemeEnum
from app.enums.preferences.font_size_enum import FontSizeEnum
from app.enums.preferences.layout_density_enum import LayoutDensityEnum


class DisplayPreferences(BaseModel):
    """Schema for user display preferences."""

    font_size: FontSizeEnum = Field(..., description="Font size preference")
    color_scheme: ColorSchemeEnum = Field(..., description="Color scheme preference")
    layout_density: LayoutDensityEnum = Field(
        ..., description="Layout density preference"
    )
    show_images: bool = Field(..., description="Whether to show images")
    compact_mode: bool = Field(..., description="Whether compact mode is enabled")
