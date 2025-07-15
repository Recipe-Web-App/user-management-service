"""Font size enumeration for user preferences."""

from enum import Enum


class FontSizeEnum(str, Enum):
    """Enum for font size preferences."""

    SMALL = "SMALL"
    MEDIUM = "MEDIUM"
    LARGE = "LARGE"
    EXTRA_LARGE = "EXTRA_LARGE"
