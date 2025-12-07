"""Display preferences schema definitions."""

from pydantic import BaseModel, Field

from app.enums.preferences.color_scheme_enum import ColorSchemeEnum
from app.enums.preferences.font_size_enum import FontSizeEnum
from app.enums.preferences.layout_density_enum import LayoutDensityEnum


class DisplayPreferences(BaseModel):
    """Schema for user display preferences."""

    font_size: FontSizeEnum | None = Field(None, description="Font size preference")
    color_scheme: ColorSchemeEnum | None = Field(
        None, description="Color scheme preference"
    )
    layout_density: LayoutDensityEnum | None = Field(
        None, description="Layout density preference"
    )
    show_images: bool | None = Field(None, description="Whether to show images")
    compact_mode: bool | None = Field(
        None, description="Whether compact mode is enabled"
    )
