"""Authentication middleware.

Provides middleware to verify JWT tokens and extract user information from
authentication headers.
"""

from typing import Annotated

from fastapi import Header, HTTPException, status
from jose import JWTError, jwt

from app.core.config import settings
from app.core.logging import get_logger

_log = get_logger(__name__)


async def get_current_user_id(
    authorization: Annotated[str, Header()],
) -> str:
    """Extract user ID from JWT token in Authorization header.

    Args:
        authorization: Authorization header containing Bearer token

    Returns:
        str: User ID from the token

    Raises:
        HTTPException: If token is invalid or missing
    """
    if not authorization.startswith("Bearer "):
        _log.warning("Invalid authorization header format")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid authorization header",
        )

    token = authorization.replace("Bearer ", "")

    try:
        payload = jwt.decode(
            token, settings.jwt_secret_key, algorithms=[settings.jwt_signing_algorithm]
        )
        user_id = payload.get("sub")
        if user_id is None:
            _log.warning("Token missing user ID")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid token",
            )
    except JWTError as e:
        _log.warning(f"JWT decode error: {e}")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token",
        ) from e
    else:
        return user_id


async def get_optional_user_id(
    authorization: Annotated[str | None, Header()] = None,
) -> str | None:
    """Extract user ID from JWT token if present, otherwise return None.

    Args:
        authorization: Authorization header containing Bearer token (optional)

    Returns:
        str | None: User ID from the token, or None if no valid token
    """
    if not authorization or not authorization.startswith("Bearer "):
        return None

    token = authorization.replace("Bearer ", "")

    try:
        payload = jwt.decode(
            token, settings.jwt_secret_key, algorithms=[settings.jwt_signing_algorithm]
        )
        user_id = payload.get("sub")
    except JWTError:
        # Silently return None for invalid tokens in optional auth
        return None
    else:
        return user_id
