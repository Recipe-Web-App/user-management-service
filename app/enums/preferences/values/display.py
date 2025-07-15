"""Display preference values."""

from enum import Enum


class FontSize(str, Enum):
    """Font size preference values."""

    SMALL = "small"
    MEDIUM = "medium"
    LARGE = "large"
    EXTRA_LARGE = "extra_large"


class ColorScheme(str, Enum):
    """Color scheme preference values."""

    LIGHT = "light"
    DARK = "dark"
    AUTO = "auto"
    HIGH_CONTRAST = "high_contrast"


class LayoutDensity(str, Enum):
    """Layout density preference values."""

    COMPACT = "compact"
    COMFORTABLE = "comfortable"
    SPACIOUS = "spacious"
