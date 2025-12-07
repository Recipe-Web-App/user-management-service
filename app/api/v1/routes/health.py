"""Health check route handler.

Defines endpoints to verify the health and status of the API service.
"""

from fastapi import APIRouter, HTTPException, status

from app.api.v1.schemas.response.liveness_response import LivenessResponse
from app.api.v1.schemas.response.readiness_response import ReadinessResponse
from app.services.health_service import health_service

router = APIRouter()


@router.get(
    "/user-management/health",
    tags=["health"],
    summary="Readiness check",
    description=(
        "Returns a 200 OK response if the server is ready to serve requests."
        "Returns degraded status (200) when database is down but Redis is healthy."
        "Returns 503 only when Redis is unavailable since JWT sessions are critical."
    ),
    response_model=ReadinessResponse,
)
async def readiness_check() -> ReadinessResponse:
    """Readiness check handler - checks app and all dependencies.

    Returns 200 OK with degraded status when only database is unavailable,
    allowing Kubernetes deployments to succeed while database reconnection
    attempts continue in the background.

    Returns:
        ReadinessResponse: Readiness status with dependency details

    Raises:
        HTTPException: 503 only if Redis is unavailable
    """
    readiness_status = await health_service.get_readiness_status()

    if not readiness_status.ready:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail=readiness_status.model_dump(),
        )

    return readiness_status


@router.get(
    "/user-management/live",
    tags=["health"],
    summary="Liveness check",
    description="Returns a 200 OK response indicating the server is alive.",
    response_model=LivenessResponse,
)
async def liveness_check() -> LivenessResponse:
    """Liveness check handler - simple check that app is alive.

    Returns:
        LivenessResponse: Liveness status
    """
    return await health_service.get_liveness_status()
