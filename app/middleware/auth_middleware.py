"""Authentication middleware.

Provides middleware to verify JWT tokens and extract user information from
authentication headers. Supports both legacy JWT validation and OAuth2 integration.
"""

from typing import Annotated

import httpx
from fastapi import Header, HTTPException, status
from jose import JWTError, jwt

from app.api.v1.schemas.downstream.auth import (
    JWTTokenPayload,
    OAuth2IntrospectionData,
)
from app.core.config import settings
from app.core.logging import get_logger

_log = get_logger(__name__)


async def _validate_jwt_token(token: str) -> JWTTokenPayload:
    """Validate JWT token using shared secret.

    Args:
        token: JWT token to validate

    Returns:
        JWTTokenPayload: Token payload data

    Raises:
        JWTError: If token validation fails
    """
    secret = settings.get_effective_jwt_secret()
    algorithm = settings.jwt_signing_algorithm

    payload_dict = jwt.decode(token, secret, algorithms=[algorithm])
    payload = JWTTokenPayload(**payload_dict)

    # Validate token type if present (OAuth2 tokens include this)
    if payload.type and payload.type != "access_token":
        _log.warning(f"Invalid token type: {payload.type}")
        raise JWTError(f"Invalid token type: {payload.type}")

    # Validate issuer if OAuth2 is enabled
    if settings.oauth2_service_enabled and not payload.iss:
        _log.warning("Token missing issuer")
        raise JWTError("Token missing issuer")

    return payload


async def _introspect_token(token: str) -> OAuth2IntrospectionData:
    """Introspect token using OAuth2 introspection endpoint.

    Args:
        token: Token to introspect

    Returns:
        OAuth2IntrospectionData: Introspection response data

    Raises:
        HTTPException: If introspection fails or token is invalid
    """
    if not settings.oauth2_introspection_url:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="OAuth2 introspection not configured",
        )

    async with httpx.AsyncClient() as client:
        try:
            response = await client.post(
                settings.oauth2_introspection_url,
                data={"token": token, "token_type_hint": "access_token"},
                auth=(settings.oauth2_client_id, settings.oauth2_client_secret),
                headers={"Content-Type": "application/x-www-form-urlencoded"},
            )
            response.raise_for_status()

            introspection_dict = response.json()
            introspection_data = OAuth2IntrospectionData(**introspection_dict)

            if not introspection_data.active:
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Token is not active",
                )
        except httpx.HTTPError as e:
            _log.error(f"OAuth2 introspection failed: {e}")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Token introspection failed",
            ) from e
        else:
            return introspection_data


async def _validate_token(token: str) -> tuple[str, list[str], str | None]:
    """Validate token and extract user ID and scopes.

    Args:
        token: JWT or opaque token to validate

    Returns:
        tuple: (user_id, scopes, client_id)

    Raises:
        HTTPException: If token validation fails
    """
    try:
        if settings.oauth2_service_enabled and settings.oauth2_introspection_enabled:
            # Use OAuth2 introspection
            introspection_data = await _introspect_token(token)
            user_id = introspection_data.user_id
            scopes = introspection_data.scopes
            client_id = introspection_data.client_id
        else:
            # Use JWT validation
            payload = await _validate_jwt_token(token)
            user_id = payload.effective_user_id
            scopes = payload.effective_scopes
            client_id = payload.client_id

        if not user_id:
            _log.warning("Token missing user identification")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token"
            )
    except JWTError as e:
        _log.warning(f"Token validation error: {e}")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token"
        ) from e
    else:
        return user_id, scopes, client_id


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
    user_id, _, _ = await _validate_token(token)
    return user_id


async def get_current_user_id_and_scopes(
    authorization: Annotated[str, Header()],
) -> tuple[str, list[str], str | None]:
    """Extract user ID, scopes, and client ID from JWT token in Authorization header.

    Args:
        authorization: Authorization header containing Bearer token

    Returns:
        tuple: (user_id, scopes, client_id)

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
    user_id, scopes, client_id = await _validate_token(token)
    return user_id, scopes, client_id


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
        user_id, _, _ = await _validate_token(token)
    except HTTPException:
        # Silently return None for invalid tokens in optional auth
        return None
    else:
        return user_id
