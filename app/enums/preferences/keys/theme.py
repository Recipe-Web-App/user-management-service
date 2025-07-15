"""Theme preference keys."""

from enum import Enum


class ThemePreferenceKey(str, Enum):
    """Theme preference keys."""

    DARK_MODE = "dark_mode"
    LIGHT_MODE = "light_mode"
    AUTO_THEME = "auto_theme"
    CUSTOM_THEME = "custom_theme"
