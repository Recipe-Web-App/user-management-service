"""Social preference keys."""

from enum import Enum


class SocialPreferenceKey(str, Enum):
    """Social preference keys."""

    FRIEND_REQUESTS = "friend_requests"
    MESSAGE_NOTIFICATIONS = "message_notifications"
    GROUP_INVITES = "group_invites"
    SHARE_ACTIVITY = "share_activity"
