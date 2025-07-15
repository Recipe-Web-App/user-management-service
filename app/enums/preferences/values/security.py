"""Security preference values."""

from enum import Enum


class SessionTimeout(int, Enum):
    """Session timeout preference values (in minutes)."""

    FIFTEEN_MINUTES = 15
    THIRTY_MINUTES = 30
    ONE_HOUR = 60
    TWO_HOURS = 120
    FOUR_HOURS = 240
    EIGHT_HOURS = 480
    TWELVE_HOURS = 720
    TWENTY_FOUR_HOURS = 1440


class PasswordStrength(str, Enum):
    """Password strength preference values."""

    WEAK = "weak"
    MEDIUM = "medium"
    STRONG = "strong"
    VERY_STRONG = "very_strong"
