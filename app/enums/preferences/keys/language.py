"""Language preference keys."""

from enum import Enum


class LanguagePreferenceKey(str, Enum):
    """Language preference keys."""

    PRIMARY_LANGUAGE = "primary_language"
    SECONDARY_LANGUAGE = "secondary_language"
    TRANSLATION_ENABLED = "translation_enabled"
