"""Theme preference values."""

from enum import Enum


class Theme(str, Enum):
    """Theme preference values."""

    LIGHT = "light"
    DARK = "dark"
    AUTO = "auto"
    CUSTOM = "custom"
