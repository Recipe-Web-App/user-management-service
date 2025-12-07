"""OAuth2 client for service-to-service authentication.

Implements OAuth2 client credentials flow for authenticating with the OAuth2 service and
other services that require OAuth2 authentication.
"""

import time
from typing import Any

import httpx
from fastapi import HTTPException, status

from app.api.v1.schemas.downstream.auth import OAuth2TokenData
from app.core.config import settings
from app.core.logging import get_logger

_log = get_logger(__name__)


class OAuth2TokenCache:
    """Simple in-memory cache for OAuth2 tokens."""

    def __init__(self) -> None:
        """Initialize token cache."""
        self._tokens: dict[str, dict[str, Any]] = {}

    def get_token(self, cache_key: str) -> dict[str, Any] | None:
        """Get cached token if valid.

        Args:
            cache_key: Cache key for the token

        Returns:
            Token data if valid, None otherwise
        """
        token_data = self._tokens.get(cache_key)
        if not token_data:
            return None

        # Check if token is expired (with 30s buffer)
        expires_at = token_data.get("expires_at", 0)
        if time.time() >= expires_at - 30:
            del self._tokens[cache_key]
            return None

        return token_data

    def set_token(self, cache_key: str, token_data: dict[str, Any]) -> None:
        """Cache token data.

        Args:
            cache_key: Cache key for the token
            token_data: Token data to cache
        """
        # Calculate expiration time
        expires_in = token_data.get("expires_in", 3600)
        expires_at = time.time() + expires_in

        cached_data = token_data.copy()
        cached_data["expires_at"] = expires_at

        self._tokens[cache_key] = cached_data
        _log.debug(f"Cached OAuth2 token for {cache_key}, expires at {expires_at}")

    def clear_tokens(self) -> None:
        """Clear all cached tokens."""
        self._tokens = {}
        _log.debug("Cleared all cached OAuth2 tokens")


class OAuth2Client:
    """OAuth2 client for service-to-service authentication."""

    def __init__(self) -> None:
        """Initialize OAuth2 client."""
        self._token_cache = OAuth2TokenCache()

    async def get_client_credentials_token(
        self, scopes: list[str] | None = None
    ) -> str:
        """Get access token using client credentials flow.

        Args:
            scopes: List of scopes to request (defaults to OAuth2_DEFAULT_SCOPES)

        Returns:
            Access token string

        Raises:
            HTTPException: If token request fails
        """
        if (
            not settings.oauth2_service_enabled
            or not settings.oauth2_service_to_service_enabled
        ):
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="OAuth2 service-to-service authentication not enabled",
            )

        if (
            not settings.oauth2_token_url
            or not settings.oauth2_client_id
            or not settings.oauth2_client_secret
        ):
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="OAuth2 client credentials not properly configured",
            )

        # Use default scopes if none provided
        if scopes is None:
            scopes = settings.oauth2_default_scopes

        # Create cache key based on scopes
        scope_str = " ".join(sorted(scopes))
        cache_key = f"client_credentials:{settings.oauth2_client_id}:{scope_str}"

        # Check cache first
        cached_token = self._token_cache.get_token(cache_key)
        if cached_token:
            _log.debug("Using cached OAuth2 client credentials token")
            return cached_token["access_token"]

        # Request new token
        _log.debug(f"Requesting OAuth2 client credentials token with scopes: {scopes}")

        async with httpx.AsyncClient() as client:
            try:
                response = await client.post(
                    settings.oauth2_token_url,
                    data={"grant_type": "client_credentials", "scope": scope_str},
                    auth=(settings.oauth2_client_id, settings.oauth2_client_secret),
                    headers={"Content-Type": "application/x-www-form-urlencoded"},
                )
                response.raise_for_status()

                token_dict = response.json()
                token_data = OAuth2TokenData(**token_dict)

                # Cache the token (convert back to dict for caching)
                self._token_cache.set_token(cache_key, token_dict)

                _log.info("Successfully obtained OAuth2 client credentials token")
            except httpx.HTTPStatusError as e:
                _log.error(f"OAuth2 client credentials request failed: {e}")
                try:
                    error_data = e.response.json()
                    error_detail = error_data.get("error_description", str(e))
                except Exception:
                    error_detail = e.response.text or str(e)

                raise HTTPException(
                    status_code=e.response.status_code,
                    detail=f"OAuth2 client credentials failed: {error_detail}",
                ) from e
            except httpx.HTTPError as e:
                _log.error(f"OAuth2 client credentials request failed: {e}")
                error_detail = str(e)
                raise HTTPException(
                    status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                    detail=f"OAuth2 client credentials failed: {error_detail}",
                ) from e
            else:
                return token_data.access_token

    async def get_service_token(self, service_scopes: list[str] | None = None) -> str:
        """Get service token for making authenticated requests to other services.

        Args:
            service_scopes: Specific scopes for the service
                (defaults to user management scopes)

        Returns:
            Bearer token string (including 'Bearer ' prefix)
        """
        if service_scopes is None:
            service_scopes = settings.oauth2_user_management_scopes

        access_token = await self.get_client_credentials_token(service_scopes)
        return f"Bearer {access_token}"

    def clear_cache(self) -> None:
        """Clear all cached tokens."""
        # Clear all cached tokens by reassigning
        self._token_cache.clear_tokens()
        _log.info("Cleared OAuth2 token cache")


# Global OAuth2 client instance
oauth2_client = OAuth2Client()


async def get_oauth2_client() -> OAuth2Client:
    """Get OAuth2 client instance.

    Returns:
        OAuth2Client: Configured OAuth2 client
    """
    return oauth2_client
