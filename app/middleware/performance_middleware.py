"""Performance monitoring middleware for tracking request metrics."""

import logging
import time
from collections.abc import Callable

from fastapi import Request, Response
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.types import ASGIApp

logger = logging.getLogger(__name__)


class PerformanceMiddleware(BaseHTTPMiddleware):
    """Middleware to track request performance metrics."""

    def __init__(self, app: ASGIApp, enable_metrics: bool = True) -> None:
        """Initialize performance middleware.

        Args:
            app: ASGI application instance
            enable_metrics: Whether to enable performance metrics collection
        """
        super().__init__(app)
        self.enable_metrics = enable_metrics
        self.request_count = 0
        self.total_response_time = 0.0

    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        """Process request and collect performance metrics."""
        if not self.enable_metrics:
            return await call_next(request)

        # Record request start time
        start_time = time.time()
        self.request_count += 1

        # Process request
        response = await call_next(request)

        # Calculate response time
        process_time = time.time() - start_time
        self.total_response_time += process_time

        # Add performance headers
        response.headers["X-Process-Time"] = str(round(process_time, 4))
        response.headers["X-Request-ID"] = request.headers.get(
            "X-Request-ID", "unknown"
        )

        # Log performance metrics
        logger.info(
            "Request processed",
            extra={
                "method": request.method,
                "url": str(request.url),
                "status_code": response.status_code,
                "process_time": process_time,
                "request_id": request.headers.get("X-Request-ID"),
            },
        )

        # Log slow requests (>1 second)
        if process_time > 1.0:
            logger.warning(
                "Slow request detected",
                extra={
                    "method": request.method,
                    "url": str(request.url),
                    "process_time": process_time,
                    "status_code": response.status_code,
                },
            )

        return response

    def get_average_response_time(self) -> float:
        """Get average response time across all requests."""
        if self.request_count == 0:
            return 0.0
        return self.total_response_time / self.request_count

    def get_metrics_summary(self) -> dict[str, float | int]:
        """Get performance metrics summary."""
        return {
            "total_requests": self.request_count,
            "total_response_time": round(self.total_response_time, 4),
            "average_response_time": round(self.get_average_response_time(), 4),
        }
