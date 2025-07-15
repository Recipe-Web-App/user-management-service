"""Profile visibility enumeration for user preferences."""

from enum import Enum


class ProfileVisibilityEnum(str, Enum):
    """Enum for profile visibility preferences."""

    PUBLIC = "PUBLIC"
    FRIENDS_ONLY = "FRIENDS_ONLY"
    PRIVATE = "PRIVATE"
