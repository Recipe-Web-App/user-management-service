"""Security preference keys."""

from enum import Enum


class SecurityPreferenceKey(str, Enum):
    """Security preference keys."""

    TWO_FACTOR_AUTH = "two_factor_auth"
    LOGIN_NOTIFICATIONS = "login_notifications"
    SESSION_TIMEOUT = "session_timeout"
    PASSWORD_REQUIREMENTS = "password_requirements"  # nosec B105
