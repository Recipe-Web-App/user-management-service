"""Theme enumeration for user preferences."""

from enum import Enum


class ThemeEnum(str, Enum):
    """Enum for theme preferences."""

    LIGHT = "LIGHT"
    DARK = "DARK"
    AUTO = "AUTO"
    CUSTOM = "CUSTOM"
