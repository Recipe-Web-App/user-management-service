"""Sound preference keys."""

from enum import Enum


class SoundPreferenceKey(str, Enum):
    """Sound preference keys."""

    NOTIFICATION_SOUNDS = "notification_sounds"
    SYSTEM_SOUNDS = "system_sounds"
    VOLUME_LEVEL = "volume_level"
    MUTE_NOTIFICATIONS = "mute_notifications"
