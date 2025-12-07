"""Base HTTP client for OAuth2-authenticated service-to-service communication."""

from typing import Any

import httpx

from app.clients.oauth2_token_manager import oauth2_token_manager
from app.core.logging import get_logger

_log = get_logger(__name__)


class BaseOAuth2ServiceClient:
    """Base HTTP client for service-to-service communication with OAuth2 authentication.

    This base client provides:
    - Persistent HTTP connection pooling via httpx.AsyncClient
    - Automatic OAuth2 token management and injection
    - Request/response logging
    - Error handling patterns
    - Configurable timeouts and retries

    Subclasses should extend this to implement specific service clients.
    """

    def __init__(
        self,
        base_url: str,
        scopes: list[str],
        timeout: float = 30.0,
    ) -> None:
        """Initialize the base service client.

        Args:
            base_url: Base URL for the downstream service
            scopes: OAuth2 scopes required for authentication
            timeout: Request timeout in seconds (default: 30.0)
        """
        self._base_url = base_url.rstrip("/")
        self._scopes = scopes
        self._timeout = timeout
        self._client: httpx.AsyncClient | None = None
        _log.debug(f"Initialized service client for {self._base_url}")

    async def _ensure_client(self) -> httpx.AsyncClient:
        """Ensure HTTP client is initialized.

        Returns:
            Configured httpx.AsyncClient instance
        """
        if self._client is None:
            self._client = httpx.AsyncClient(
                base_url=self._base_url,
                timeout=self._timeout,
                follow_redirects=True,
            )
            _log.debug(f"Created HTTP client for {self._base_url}")
        return self._client

    async def _get_auth_headers(self) -> dict[str, str]:
        """Get authentication headers with valid OAuth2 token.

        Returns:
            Dictionary with Authorization header

        Raises:
            httpx.HTTPError: If token acquisition fails
        """
        token = await oauth2_token_manager.get_valid_token(self._scopes)
        return {"Authorization": f"Bearer {token}"}

    async def _post(
        self,
        endpoint: str,
        json_data: dict[str, Any] | None = None,
        **kwargs: Any,
    ) -> httpx.Response:
        """Make an authenticated POST request.

        Args:
            endpoint: API endpoint path (relative to base_url)
            json_data: JSON payload to send
            **kwargs: Additional arguments to pass to httpx.post

        Returns:
            httpx.Response object

        Raises:
            httpx.HTTPError: If request fails
        """
        client = await self._ensure_client()
        auth_headers = await self._get_auth_headers()
        headers = {**auth_headers, **kwargs.pop("headers", {})}

        url = f"{self._base_url}{endpoint}"
        _log.debug(f"POST {url}")

        response = await client.post(
            endpoint,
            json=json_data,
            headers=headers,
            **kwargs,
        )

        _log.debug(f"POST {url} -> {response.status_code}")
        return response

    async def _get(
        self,
        endpoint: str,
        params: dict[str, Any] | None = None,
        **kwargs: Any,
    ) -> httpx.Response:
        """Make an authenticated GET request.

        Args:
            endpoint: API endpoint path (relative to base_url)
            params: Query parameters
            **kwargs: Additional arguments to pass to httpx.get

        Returns:
            httpx.Response object

        Raises:
            httpx.HTTPError: If request fails
        """
        client = await self._ensure_client()
        auth_headers = await self._get_auth_headers()
        headers = {**auth_headers, **kwargs.pop("headers", {})}

        url = f"{self._base_url}{endpoint}"
        _log.debug(f"GET {url}")

        response = await client.get(
            endpoint,
            params=params,
            headers=headers,
            **kwargs,
        )

        _log.debug(f"GET {url} -> {response.status_code}")
        return response

    async def close(self) -> None:
        """Close the HTTP client and clean up connections."""
        if self._client is not None:
            await self._client.aclose()
            self._client = None
            _log.debug(f"Closed HTTP client for {self._base_url}")

    async def __aenter__(self) -> "BaseOAuth2ServiceClient":
        """Context manager entry."""
        await self._ensure_client()
        return self

    async def __aexit__(self, _exc_type: Any, _exc_val: Any, _exc_tb: Any) -> None:
        """Context manager exit."""
        await self.close()
