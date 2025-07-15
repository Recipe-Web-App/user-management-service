"""Preference type enumeration for user preferences."""

from enum import Enum


class PreferenceType(str, Enum):
    """Preference type enum for categorizing user preferences."""

    NOTIFICATION = "notification"
    PRIVACY = "privacy"
    DISPLAY = "display"
    ACCESSIBILITY = "accessibility"
    LANGUAGE = "language"
    TIMEZONE = "timezone"
    THEME = "theme"
    SOUND = "sound"
    SOCIAL = "social"
    SECURITY = "security"
