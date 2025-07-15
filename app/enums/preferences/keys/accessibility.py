"""Accessibility preference keys."""

from enum import Enum


class AccessibilityPreferenceKey(str, Enum):
    """Accessibility preference keys."""

    SCREEN_READER = "screen_reader"
    HIGH_CONTRAST = "high_contrast"
    REDUCED_MOTION = "reduced_motion"
    LARGE_TEXT = "large_text"
    KEYBOARD_NAVIGATION = "keyboard_navigation"
