"""Security utilities for handling sensitive data."""

from typing import Any


class SensitiveData:
    """Wrapper for sensitive data that redacts in logs but shows in API responses."""

    def __init__(self, value: str) -> None:
        """Initialize with the sensitive value."""
        self._value = value

    def __str__(self) -> str:
        """Return redacted value for logging."""
        return "***REDACTED***"

    def __repr__(self) -> str:
        """Return redacted value for debugging."""
        return "SensitiveData(***REDACTED***)"

    def get_value(self) -> str:
        """Get the actual value for API responses."""
        return self._value

    def __eq__(self, other: Any) -> bool:
        """Compare with other values."""
        if isinstance(other, SensitiveData):
            return self._value == other._value
        return self._value == other

    def __hash__(self) -> int:
        """Make the object hashable."""
        return hash(self._value)
