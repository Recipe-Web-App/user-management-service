"""OAuth2 token manager for service-to-service authentication.

Manages OAuth2 access tokens for downstream service calls, refreshing only when expired.
"""

import asyncio
import time
from typing import Any

import httpx

from app.core.config import settings
from app.core.logging import get_logger

_log = get_logger(__name__)


class OAuth2TokenManager:
    """Manages OAuth2 tokens for service-to-service authentication.

    This manager stores tokens and their expiration timestamps, only fetching new tokens
    when the current token has expired. Unlike the general OAuth2 client cache, this
    uses exact expiration without arbitrary buffers.
    """

    def __init__(self) -> None:
        """Initialize the token manager."""
        self._tokens: dict[str, dict[str, Any]] = {}
        self._lock = asyncio.Lock()

    async def get_valid_token(self, scopes: list[str]) -> str:
        """Get a valid access token, refreshing if expired.

        Args:
            scopes: List of OAuth2 scopes required for the token

        Returns:
            Valid access token string (without Bearer prefix)

        Raises:
            httpx.HTTPError: If token request fails
        """
        scope_str = " ".join(sorted(scopes))
        cache_key = f"client_credentials:{scope_str}"

        async with self._lock:
            # Check if we have a cached token
            cached_data = self._tokens.get(cache_key)

            if cached_data:
                expires_at = cached_data.get("expires_at", 0)
                current_time = time.time()

                # Only refresh if actually expired
                if current_time < expires_at:
                    _log.debug(
                        f"Using cached token for scopes: {scopes}, "
                        f"expires in {int(expires_at - current_time)}s"
                    )
                    return cached_data["access_token"]
                _log.debug(f"Token expired for scopes: {scopes}, fetching new token")

            # Fetch new token
            token_data = await self._fetch_token(scope_str)

            # Cache the token with expiration
            expires_in = token_data.get("expires_in", 3600)
            expires_at = time.time() + expires_in

            self._tokens[cache_key] = {
                "access_token": token_data["access_token"],
                "expires_at": expires_at,
                "expires_in": expires_in,
            }

            _log.info(
                f"Obtained new OAuth2 token for scopes: {scopes}, "
                f"expires in {expires_in}s"
            )

            return token_data["access_token"]

    async def _fetch_token(self, scope_str: str) -> dict[str, Any]:
        """Fetch a new OAuth2 token using client credentials flow.

        Args:
            scope_str: Space-separated scope string

        Returns:
            Token response data from OAuth2 server

        Raises:
            httpx.HTTPError: If token request fails
            RuntimeError: If OAuth2 is not properly configured
        """
        if (
            not settings.oauth2_service_enabled
            or not settings.oauth2_service_to_service_enabled
        ):
            raise RuntimeError("OAuth2 service-to-service authentication not enabled")

        if (
            not settings.oauth2_token_url
            or not settings.oauth2_client_id
            or not settings.oauth2_client_secret
        ):
            raise RuntimeError("OAuth2 client credentials not properly configured")

        _log.debug(f"Requesting OAuth2 token from {settings.oauth2_token_url}")

        async with httpx.AsyncClient() as client:
            response = await client.post(
                settings.oauth2_token_url,
                data={
                    "grant_type": "client_credentials",
                    "scope": scope_str,
                },
                auth=(settings.oauth2_client_id, settings.oauth2_client_secret),
                headers={"Content-Type": "application/x-www-form-urlencoded"},
                timeout=10.0,
            )
            response.raise_for_status()

            return response.json()

    def clear_cache(self) -> None:
        """Clear all cached tokens."""
        self._tokens.clear()
        _log.debug("Cleared all cached OAuth2 tokens")


# Global token manager instance
oauth2_token_manager = OAuth2TokenManager()
