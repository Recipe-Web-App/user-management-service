"""Security utilities for handling sensitive data."""

from typing import Any

from jose import jwt
from sqlalchemy import String, TypeDecorator

from app.core.config import settings


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

    def get_raw_value(self) -> str:
        """Get the raw value for database operations."""
        return self._value

    def __eq__(self, other: Any) -> bool:
        """Compare with other values."""
        if isinstance(other, SensitiveData):
            return self._value == other._value
        return self._value == other

    def __hash__(self) -> int:
        """Make the object hashable."""
        return hash(self._value)

    def __getitem__(self, key: str) -> Any:
        """Support dictionary-style access for Pydantic serialization."""
        if key == "value":
            return self._value
        raise KeyError(key)

    def keys(self) -> list[str]:
        """Return keys for dictionary-style access."""
        return ["value"]


class SecurePasswordHash(TypeDecorator):
    """SQLAlchemy type for securely storing password hashes.

    This type automatically wraps password hashes in SensitiveData to prevent accidental
    exposure in logs and serialization.
    """

    impl = String
    cache_ok = True

    def __init__(self, length: int = 255) -> None:
        """Initialize with specified length."""
        super().__init__(length)

    def process_bind_param(
        self, value: str | SensitiveData | None, _dialect: Any
    ) -> str | None:
        """Process value when binding to database."""
        if value is None:
            return None
        if isinstance(value, SensitiveData):
            return value.get_raw_value()
        return str(value)

    def process_result_value(
        self, value: str | None, _dialect: Any
    ) -> SensitiveData | None:
        """Process value when retrieving from database."""
        if value is None:
            return None
        return SensitiveData(value)

    def process_literal_param(
        self, value: str | SensitiveData | None, _dialect: Any
    ) -> str:
        """Process literal parameter."""
        if value is None:
            return ""
        if isinstance(value, SensitiveData):
            return value.get_raw_value()
        return str(value)


def decode_jwt_token(token: str) -> dict:
    """Decode and validate JWT token.

    Args:
        token: JWT token to decode

    Returns:
        dict: Decoded token payload

    Raises:
        jose.JWTError: If token is invalid or expired
    """
    return jwt.decode(
        token,
        settings.jwt_secret_key,
        algorithms=[settings.jwt_signing_algorithm],
    )
