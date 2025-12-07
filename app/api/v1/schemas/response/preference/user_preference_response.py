"""User preference response schema definitions."""

from app.api.v1.schemas.base_schema_model import BaseSchemaModel
from app.api.v1.schemas.common.accessibility_preferences import AccessibilityPreferences
from app.api.v1.schemas.common.display_preferences import DisplayPreferences
from app.api.v1.schemas.common.language_preferences import LanguagePreferences
from app.api.v1.schemas.common.notification_preferences import NotificationPreferences
from app.api.v1.schemas.common.privacy_preferences import PrivacyPreferences
from app.api.v1.schemas.common.security_preferences import SecurityPreferences
from app.api.v1.schemas.common.social_preferences import SocialPreferences
from app.api.v1.schemas.common.sound_preferences import SoundPreferences
from app.api.v1.schemas.common.theme_preferences import ThemePreferences


class UserPreferenceResponse(BaseSchemaModel):
    """Response schema for all user preferences, grouped by category."""

    user_id: str
    notification: NotificationPreferences
    display: DisplayPreferences
    theme: ThemePreferences
    privacy: PrivacyPreferences
    security: SecurityPreferences
    sound: SoundPreferences
    social: SocialPreferences
    language: LanguagePreferences
    accessibility: AccessibilityPreferences
