"""Common schemas for the API."""

from app.api.v1.schemas.common.accessibility_preferences import AccessibilityPreferences
from app.api.v1.schemas.common.display_preferences import DisplayPreferences
from app.api.v1.schemas.common.language_preferences import LanguagePreferences
from app.api.v1.schemas.common.notification import Notification
from app.api.v1.schemas.common.notification_preferences import NotificationPreferences
from app.api.v1.schemas.common.privacy_preferences import PrivacyPreferences
from app.api.v1.schemas.common.security_preferences import SecurityPreferences
from app.api.v1.schemas.common.social_preferences import SocialPreferences
from app.api.v1.schemas.common.sound_preferences import SoundPreferences
from app.api.v1.schemas.common.theme_preferences import ThemePreferences
from app.api.v1.schemas.common.token import Token
from app.api.v1.schemas.common.user import User

__all__ = [
    "AccessibilityPreferences",
    "DisplayPreferences",
    "LanguagePreferences",
    "Notification",
    "NotificationPreferences",
    "PrivacyPreferences",
    "SecurityPreferences",
    "SocialPreferences",
    "SoundPreferences",
    "ThemePreferences",
    "Token",
    "User",
]
