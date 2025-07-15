"""Sound preference values."""

from enum import Enum


class VolumeLevel(str, Enum):
    """Volume level preference values."""

    MUTED = "muted"
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
