"""Privacy preference values."""

from enum import Enum


class ProfileVisibility(str, Enum):
    """Profile visibility preference values."""

    PUBLIC = "public"
    FRIENDS_ONLY = "friends_only"
    PRIVATE = "private"
