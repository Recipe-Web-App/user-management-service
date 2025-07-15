"""Color scheme enumeration for user preferences."""

from enum import Enum


class ColorSchemeEnum(str, Enum):
    """Enum for color scheme preferences."""

    LIGHT = "LIGHT"
    DARK = "DARK"
    AUTO = "AUTO"
    HIGH_CONTRAST = "HIGH_CONTRAST"
