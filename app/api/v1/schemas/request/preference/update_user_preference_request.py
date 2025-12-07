"""Update user preference request schema."""

from pydantic import BaseModel, Field

from app.api.v1.schemas.common.accessibility_preferences import AccessibilityPreferences
from app.api.v1.schemas.common.display_preferences import DisplayPreferences
from app.api.v1.schemas.common.language_preferences import LanguagePreferences
from app.api.v1.schemas.common.notification_preferences import NotificationPreferences
from app.api.v1.schemas.common.privacy_preferences import PrivacyPreferences
from app.api.v1.schemas.common.security_preferences import SecurityPreferences
from app.api.v1.schemas.common.social_preferences import SocialPreferences
from app.api.v1.schemas.common.sound_preferences import SoundPreferences
from app.api.v1.schemas.common.theme_preferences import ThemePreferences


class UpdateUserPreferenceRequest(BaseModel):
    """Request schema for updating user preferences across all categories.

    All fields are optional. Only provided fields will be updated; fields set to None or
    omitted will leave the corresponding database values unchanged.
    """

    accessibility: AccessibilityPreferences | None = Field(
        None, description="Accessibility preferences update"
    )
    display: DisplayPreferences | None = Field(
        None, description="Display preferences update"
    )
    language: LanguagePreferences | None = Field(
        None, description="Language preferences update"
    )
    notification: NotificationPreferences | None = Field(
        None, description="Notification preferences update"
    )
    privacy: PrivacyPreferences | None = Field(
        None, description="Privacy preferences update"
    )
    security: SecurityPreferences | None = Field(
        None, description="Security preferences update"
    )
    social: SocialPreferences | None = Field(
        None, description="Social preferences update"
    )
    sound: SoundPreferences | None = Field(None, description="Sound preferences update")
    theme: ThemePreferences | None = Field(None, description="Theme preferences update")
