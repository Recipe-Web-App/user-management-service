"""Request ID middleware.

Provides middleware to assign unique request IDs to each incoming HTTP request for
traceability.
"""

import contextlib
import uuid
from collections.abc import Awaitable, Callable, Iterator
from typing import Any

from fastapi import Request, Response
from starlette.middleware.base import BaseHTTPMiddleware

from app.core.logging import set_request_id

REQUEST_ID_HEADER = "X-Request-ID"


class RequestIDMiddleware(BaseHTTPMiddleware):
    """Middleware to add request ID to all requests."""

    async def dispatch(
        self, request: Request, call_next: Callable[[Request], Awaitable[Response]]
    ) -> Response:
        """Process the request and add request ID."""
        request_id = request.headers.get(REQUEST_ID_HEADER, str(uuid.uuid4()))
        request.state.request_id = request_id
        # Use context manager to ensure request_id is available throughout the request
        with _request_context(request_id):
            response = await call_next(request)
            response.headers[REQUEST_ID_HEADER] = request_id
            return response


@contextlib.contextmanager
def _request_context(request_id: str) -> Iterator[None]:
    """Context manager to set request ID for the duration of a request."""
    set_request_id(request_id)
    try:
        yield
    finally:
        # Reset to default when context exits
        set_request_id("NULL")


class CustomRequestIDMiddleware:
    """Custom middleware that preserves contextvars propagation."""

    def __init__(self, app: Any) -> None:
        """Initialize the custom middleware.

        Args:
            app: The ASGI application to wrap.

        """
        self.app = app

    async def __call__(
        self,
        scope: dict[str, Any],
        receive: Callable[[], Awaitable[dict[str, Any]]],
        send: Callable[[dict[str, Any]], Awaitable[None]],
    ) -> None:
        """Process the ASGI request and add request ID.

        Args:
            scope: The ASGI scope.
            receive: The receive callable.
            send: The send callable.

        """
        if scope["type"] == "http":
            request = Request(scope, receive)
            request_id = request.headers.get(REQUEST_ID_HEADER, str(uuid.uuid4()))
            request.state.request_id = request_id
            # Use context manager to ensure request_id is available throughout
            # the request
            with _request_context(request_id):

                # Create a custom send function to add headers
                async def custom_send(message: dict[str, Any]) -> None:
                    if message["type"] == "http.response.start":
                        message["headers"] = list(message.get("headers", []))
                        header_tuple = (
                            REQUEST_ID_HEADER.encode(),
                            request_id.encode(),
                        )
                        message["headers"].append(header_tuple)
                    await send(message)

                await self.app(scope, receive, custom_send)
        else:
            await self.app(scope, receive, send)


async def request_id_middleware(
    request: Request, call_next: Callable[[Request], Awaitable[Response]]
) -> Response:
    """Function-based middleware to add request ID to all requests."""
    request_id = request.headers.get(REQUEST_ID_HEADER, str(uuid.uuid4()))
    request.state.request_id = request_id
    # Use context manager to ensure request_id is available throughout the request
    with _request_context(request_id):
        response = await call_next(request)
        response.headers[REQUEST_ID_HEADER] = request_id
        return response
