"""Layout density enumeration for user preferences."""

from enum import Enum


class LayoutDensityEnum(str, Enum):
    """Enum for layout density preferences."""

    COMPACT = "COMPACT"
    COMFORTABLE = "COMFORTABLE"
    SPACIOUS = "SPACIOUS"
