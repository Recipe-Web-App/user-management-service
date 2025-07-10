"""Constants and configuration values for the application."""

from dataclasses import dataclass
from typing import ClassVar


@dataclass(frozen=True)
class Constants:
    """Application constants and configuration values."""

    MIN_PASSWORD_LENGTH: ClassVar[int] = 8
    MAX_PASSWORD_LENGTH: ClassVar[int] = 128

    _instance: ClassVar["Constants | None"] = None

    def __new__(cls) -> "Constants":
        """Singleton pattern - return the same instance."""
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance


CONSTANTS = Constants()
