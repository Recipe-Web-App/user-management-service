"""Token type enumeration for authentication."""

from enum import Enum


class TokenType(str, Enum):
    """Token type enum."""

    BEARER = "bearer"
