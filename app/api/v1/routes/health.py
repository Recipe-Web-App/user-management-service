"""Health check route handler.

Defines endpoints to verify the health and status of the API service.
"""

from fastapi import APIRouter
from fastapi.responses import JSONResponse

router = APIRouter()


@router.get(
    "/user-management/health",
    tags=["health"],
    summary="Health check",
    description="Returns a 200 OK response indicating the server is up.",
    response_class=JSONResponse,
)
async def health_check() -> JSONResponse:
    """Health check handler.

    Returns:
        JSONResponse: OK
    """
    content = {"status": "ok"}
    return JSONResponse(content=content, status_code=200)
