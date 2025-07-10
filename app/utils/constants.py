"""Constants and configuration values for the application."""

from dataclasses import dataclass
from typing import ClassVar


@dataclass(frozen=True)
class Constants:
    """Application constants and configuration values."""

    MIN_PASSWORD_LENGTH: ClassVar[int] = 8
    MAX_PASSWORD_LENGTH: ClassVar[int] = 128

    def __new__(cls) -> "Constants":
        """Singleton pattern - return the same instance."""
        _ = cls  # Avoids vulture error

        return CONSTANTS


CONSTANTS = Constants()
