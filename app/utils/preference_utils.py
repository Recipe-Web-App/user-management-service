"""Utility functions for user preferences."""

from typing import Any
from uuid import UUID

from app.db.sql.models.user.user_preferences import UserPreferences
from app.enums.preferences import (
    AccessibilityPreferenceKey,
    ColorScheme,
    DisplayPreferenceKey,
    FontSize,
    Language,
    LanguagePreferenceKey,
    LayoutDensity,
    NotificationPreferenceKey,
    PasswordStrength,
    PreferenceType,
    PrivacyPreferenceKey,
    ProfileVisibility,
    SecurityPreferenceKey,
    SessionTimeout,
    SocialPreferenceKey,
    SoundPreferenceKey,
    Theme,
    ThemePreferenceKey,
    VolumeLevel,
)


def get_preference_key_enum(preference_type: PreferenceType) -> type[Any]:
    """Get the appropriate preference key enum for a given preference type.

    Args:
        preference_type: The type of preference

    Returns:
        type[Any]: The enum class for the preference keys

    Raises:
        ValueError: If the preference type is not supported
    """
    key_enums = {
        PreferenceType.NOTIFICATION: NotificationPreferenceKey,
        PreferenceType.PRIVACY: PrivacyPreferenceKey,
        PreferenceType.DISPLAY: DisplayPreferenceKey,
        PreferenceType.ACCESSIBILITY: AccessibilityPreferenceKey,
        PreferenceType.LANGUAGE: LanguagePreferenceKey,
        PreferenceType.THEME: ThemePreferenceKey,
        PreferenceType.SOUND: SoundPreferenceKey,
        PreferenceType.SOCIAL: SocialPreferenceKey,
        PreferenceType.SECURITY: SecurityPreferenceKey,
    }

    if preference_type not in key_enums:
        raise ValueError(f"Unsupported preference type: {preference_type}")

    return key_enums[preference_type]


def get_preference_value_enum(preference_key: str) -> type[Any] | None:
    """Get the appropriate preference value enum for a given preference key.

    Args:
        preference_key: The preference key

    Returns:
        type[Any] | None: The enum class for the preference values, or None if
        not applicable
    """
    # Map preference keys to their value enums
    value_enums = {
        # Display preferences
        DisplayPreferenceKey.FONT_SIZE.value: FontSize,
        DisplayPreferenceKey.COLOR_SCHEME.value: ColorScheme,
        DisplayPreferenceKey.LAYOUT_DENSITY.value: LayoutDensity,
        # Privacy preferences
        PrivacyPreferenceKey.PROFILE_VISIBILITY.value: ProfileVisibility,
        # Language preferences
        LanguagePreferenceKey.PRIMARY_LANGUAGE.value: Language,
        LanguagePreferenceKey.SECONDARY_LANGUAGE.value: Language,
        # Theme preferences
        ThemePreferenceKey.DARK_MODE.value: Theme,
        ThemePreferenceKey.LIGHT_MODE.value: Theme,
        ThemePreferenceKey.AUTO_THEME.value: Theme,
        ThemePreferenceKey.CUSTOM_THEME.value: Theme,
        # Sound preferences
        SoundPreferenceKey.VOLUME_LEVEL.value: VolumeLevel,
        # Security preferences
        SecurityPreferenceKey.SESSION_TIMEOUT.value: SessionTimeout,
        SecurityPreferenceKey.PASSWORD_REQUIREMENTS.value: PasswordStrength,
    }

    return value_enums.get(preference_key)


def validate_preference_value(preference_key: str, value: Any) -> bool:
    """Validate a preference value against its expected enum.

    Args:
        preference_key: The preference key
        value: The value to validate

    Returns:
        bool: True if the value is valid, False otherwise
    """
    value_enum = get_preference_value_enum(preference_key)
    if value_enum is None:
        # No specific enum validation for this key, accept any value
        return True

    # Check if the value is a valid enum value
    try:
        valid_values = [member.value for member in value_enum]
    except (ValueError, TypeError):
        return False
    else:
        return value in valid_values


def create_preference(
    user_id: UUID,
    preference_type: PreferenceType,
    preference_key: str,
    preference_value: Any,
    description: str | None = None,
) -> UserPreferences:
    """Create a user preference with validation.

    Args:
        user_id: The user's unique identifier
        preference_type: The type of preference
        preference_key: The preference key
        preference_value: The preference value
        description: Optional description of the preference

    Returns:
        UserPreferences: The created preference instance

    Raises:
        ValueError: If the preference type or key is invalid
    """
    # Validate preference type
    if not isinstance(preference_type, PreferenceType):
        raise TypeError(f"Invalid preference type: {preference_type}")

    # Get the appropriate key enum for validation
    key_enum = get_preference_key_enum(preference_type)

    # Validate preference key
    valid_keys = [member.value for member in key_enum]
    if preference_key not in valid_keys:
        raise ValueError(
            f"Invalid preference key '{preference_key}' for type "
            f"'{preference_type.value}'. Valid keys: {valid_keys}"
        )

    # Validate preference value if applicable
    if not validate_preference_value(preference_key, preference_value):
        value_enum = get_preference_value_enum(preference_key)
        if value_enum:
            valid_values = [val.value for val in value_enum]
            raise ValueError(
                f"Invalid preference value '{preference_value}' for key "
                f"'{preference_key}'. Valid values: {valid_values}"
            )

    return UserPreferences(
        user_id=user_id,
        preference_type=preference_type.value,
        preference_key=preference_key,
        preference_value=preference_value,
        description=description,
    )


def get_preference_type_from_key(preference_key: str) -> PreferenceType | None:
    """Get the preference type from a preference key.

    Args:
        preference_key: The preference key

    Returns:
        PreferenceType | None: The preference type, or None if not found
    """
    # Check each preference key enum to find the matching key
    key_enums = [
        (PreferenceType.NOTIFICATION, NotificationPreferenceKey),
        (PreferenceType.PRIVACY, PrivacyPreferenceKey),
        (PreferenceType.DISPLAY, DisplayPreferenceKey),
        (PreferenceType.ACCESSIBILITY, AccessibilityPreferenceKey),
        (PreferenceType.LANGUAGE, LanguagePreferenceKey),
        (PreferenceType.THEME, ThemePreferenceKey),
        (PreferenceType.SOUND, SoundPreferenceKey),
        (PreferenceType.SOCIAL, SocialPreferenceKey),
        (PreferenceType.SECURITY, SecurityPreferenceKey),
    ]

    for preference_type, key_enum in key_enums:
        valid_keys = [member.value for member in key_enum]  # type: ignore
        if preference_key in valid_keys:
            return preference_type

    return None
