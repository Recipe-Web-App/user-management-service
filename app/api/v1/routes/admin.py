"""Admin routes.

Provides endpoints for admin operations.
"""

from http import HTTPStatus
from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, Path
from fastapi.responses import JSONResponse

from app.api.v1.schemas.response.admin.redis_session_stats_response import (
    RedisSessionStatsResponse,
)
from app.api.v1.schemas.response.admin.user_stats_response import UserStatsResponse
from app.api.v1.schemas.response.error_response import ErrorResponse
from app.db.redis.redis_database_manager import get_redis_session
from app.db.redis.redis_database_session import RedisDatabaseSession
from app.db.sql.sql_database_manager import get_db
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.middleware.auth_middleware import get_current_user_id
from app.services.admin_service import AdminService

router = APIRouter()


async def get_admin_service(
    db: Annotated[SqlDatabaseSession, Depends(get_db)],
    redis_session: Annotated[RedisDatabaseSession, Depends(get_redis_session)],
) -> AdminService:
    """Get admin service instance.

    Args:
        db: Database session
        redis_session: Redis session

    Returns:
        AdminService: Admin service instance
    """
    return AdminService(db, redis_session)


@router.get(
    "/user-management/admin/redis/sessions/stats",
    tags=["admin"],
    summary="Get Redis session stats",
    description="Return Redis session statistics.",
    response_model=RedisSessionStatsResponse,
    responses={
        HTTPStatus.OK: {
            "model": RedisSessionStatsResponse,
            "description": "Session stats returned.",
        },
        HTTPStatus.UNAUTHORIZED: {
            "model": ErrorResponse,
            "description": "Invalid or missing authorization token.",
        },
        HTTPStatus.FORBIDDEN: {
            "model": ErrorResponse,
            "description": "Admin privileges required.",
        },
        HTTPStatus.INTERNAL_SERVER_ERROR: {
            "model": ErrorResponse,
            "description": "Internal server error.",
        },
        HTTPStatus.SERVICE_UNAVAILABLE: {
            "model": ErrorResponse,
            "description": "Service temporarily unavailable.",
        },
    },
)
async def get_redis_session_stats(
    user_id: Annotated[str, Depends(get_current_user_id)],
    admin_service: Annotated[AdminService, Depends(get_admin_service)],
) -> RedisSessionStatsResponse:
    """Get Redis session statistics.

    Args:
        user_id: The user ID from JWT (authorization)
        admin_service: Admin service instance

    Returns:
        RedisSessionStatsResponse: Redis session stats
    """
    return await admin_service.get_redis_session_stats(user_id)


@router.get(
    "/user-management/admin/users/stats",
    tags=["admin"],
    summary="Get user statistics",
    description="Return user statistics.",
    response_model=UserStatsResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserStatsResponse,
            "description": "User stats returned.",
        },
        HTTPStatus.UNAUTHORIZED: {
            "model": ErrorResponse,
            "description": "Invalid or missing authorization token.",
        },
        HTTPStatus.FORBIDDEN: {
            "model": ErrorResponse,
            "description": "Admin privileges required.",
        },
        HTTPStatus.INTERNAL_SERVER_ERROR: {
            "model": ErrorResponse,
            "description": "Internal server error.",
        },
    },
)
async def get_user_stats(
    user_id: Annotated[str, Depends(get_current_user_id)],
    admin_service: Annotated[AdminService, Depends(get_admin_service)],
) -> UserStatsResponse:
    """Get user statistics.

    Args:
        user_id: The user ID from JWT (authorization)
        admin_service: Admin service instance

    Returns:
        UserStatsResponse: User statistics
    """
    return await admin_service.get_user_stats(user_id)


@router.get(
    "/user-management/admin/health",
    tags=["admin"],
    summary="System health check",
    description="Return system health status.",
    response_model=dict,
    responses={
        HTTPStatus.OK: {"model": dict, "description": "System health returned."},
        HTTPStatus.INTERNAL_SERVER_ERROR: {
            "model": ErrorResponse,
            "description": "Internal server error.",
        },
    },
)
async def get_system_health(
    user_id: Annotated[str, Depends(get_current_user_id)],
    admin_service: Annotated[AdminService, Depends(get_admin_service)],
) -> JSONResponse:
    """Get system health status.

    Args:
        user_id: The user ID from JWT (authorization)
        admin_service: Admin service instance

    Returns:
        JSONResponse: System health
    """
    return JSONResponse(await admin_service.get_system_health(user_id))


@router.get(
    "/user-management/admin/activity/recent-logins",
    tags=["admin"],
    summary="Recent user logins",
    description="Return recent user logins.",
    response_model=dict,
    responses={
        HTTPStatus.OK: {"model": dict, "description": "Recent logins returned."},
        HTTPStatus.INTERNAL_SERVER_ERROR: {
            "model": ErrorResponse,
            "description": "Internal server error.",
        },
    },
)
async def get_recent_logins(
    user_id: Annotated[str, Depends(get_current_user_id)],
    admin_service: Annotated[AdminService, Depends(get_admin_service)],
) -> JSONResponse:
    """Get recent user logins.

    Args:
        user_id: The user ID from JWT (authorization)
        admin_service: Admin service instance

    Returns:
        JSONResponse: Recent logins
    """
    return JSONResponse(await admin_service.get_recent_logins(user_id))


@router.post(
    "/user-management/admin/users/{user_id}/force-logout",
    tags=["admin"],
    summary="Force logout user",
    description="Force logout a user.",
    response_model=dict,
    responses={
        HTTPStatus.OK: {"model": dict, "description": "User force-logout triggered."},
        HTTPStatus.BAD_REQUEST: {"model": ErrorResponse, "description": "Bad request."},
        HTTPStatus.INTERNAL_SERVER_ERROR: {
            "model": ErrorResponse,
            "description": "Internal server error.",
        },
    },
)
async def force_logout_user(
    user_id: Annotated[UUID, Path(description="User ID")],
    admin_user_id: Annotated[str, Depends(get_current_user_id)],
    admin_service: Annotated[AdminService, Depends(get_admin_service)],
) -> JSONResponse:
    """Force logout a user.

    Args:
        user_id: The user's unique identifier
        admin_user_id: The admin user ID from JWT (authorization)
        admin_service: Admin service instance

    Returns:
        JSONResponse: Force logout result
    """
    return JSONResponse(await admin_service.force_logout_user(admin_user_id, user_id))


@router.delete(
    "/user-management/admin/redis/sessions",
    tags=["admin"],
    summary="Clear all Redis sessions",
    description="Clear all Redis sessions.",
    response_model=dict,
    responses={
        HTTPStatus.OK: {"model": dict, "description": "All sessions cleared."},
        HTTPStatus.INTERNAL_SERVER_ERROR: {
            "model": ErrorResponse,
            "description": "Internal server error.",
        },
    },
)
async def clear_redis_sessions(
    user_id: Annotated[str, Depends(get_current_user_id)],
    admin_service: Annotated[AdminService, Depends(get_admin_service)],
) -> JSONResponse:
    """Clear all Redis sessions.

    Args:
        user_id: The user ID from JWT (authorization)
        admin_service: Admin service instance

    Returns:
        JSONResponse: Clear sessions result
    """
    return JSONResponse(await admin_service.clear_redis_sessions(user_id))
