"""Exception handlers.

Contains FastAPI exception handler functions to map exceptions to HTTP responses.
"""

import traceback
from http import HTTPStatus

from fastapi import Request
from fastapi.exceptions import HTTPException
from fastapi.responses import JSONResponse, Response

from app.core.logging import get_logger

_log = get_logger(__name__)


async def unhandled_exception_handler(_request: Request, exc: Exception) -> Response:
    """Handle unhandled exceptions in the FastAPI application.

    Args:
        request (Request): The incoming request that caused the exception.
        exc (Exception): The exception that was raised.

    Raises:
        Response: JSON response with error details.
    """
    if isinstance(exc, HTTPException):
        raise exc

    _log.error("Unhandled exception occurred: %s\n%s", str(exc), traceback.format_exc())
    return JSONResponse(
        status_code=HTTPStatus.INTERNAL_SERVER_ERROR,
        content={"detail": "An unexpected error occurred."},
    )
