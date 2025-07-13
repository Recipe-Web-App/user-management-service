"""Authentication route handler.

Defines endpoints for user authentication and account management.
"""

from http import HTTPStatus
from typing import Annotated

from fastapi import APIRouter, Body, Depends
from fastapi.responses import JSONResponse

from app.api.v1.schemas.request.user_registration_request import UserRegistrationRequest
from app.api.v1.schemas.response.error_response import ErrorResponse
from app.api.v1.schemas.response.user_registration_response import (
    UserRegistrationResponse,
)
from app.db.redis.redis_database_manager import get_redis_session
from app.db.redis.redis_database_session import RedisDatabaseSession
from app.db.sql.sql_database_manager import get_db
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.services.auth_service import AuthService

router = APIRouter()


async def get_auth_service(
    db: Annotated[SqlDatabaseSession, Depends(get_db)],
    redis_session: Annotated[RedisDatabaseSession, Depends(get_redis_session)],
) -> AuthService:
    """Get auth service instance."""
    return AuthService(db, redis_session)


@router.post(
    "/user-management/auth/register",
    tags=["authentication"],
    summary="Register a new user",
    description="Create a new user account",
    response_model=UserRegistrationResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserRegistrationResponse,
            "description": "User registered successfully",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Bad request",
        },
        HTTPStatus.UNPROCESSABLE_ENTITY: {
            "model": ErrorResponse,
            "description": "Validation error",
        },
        HTTPStatus.INTERNAL_SERVER_ERROR: {
            "model": ErrorResponse,
            "description": "Internal server error",
        },
    },
)
async def register(
    user_data: Annotated[UserRegistrationRequest, Body(..., embed=True)],
    auth_service: Annotated[AuthService, Depends(get_auth_service)],
) -> UserRegistrationResponse:
    """Register a new user.

    Args:
        user_data: User registration data
        auth_service: Auth service instance

    Returns:
        UserRegistrationResponse: Registration result with user data and access token

    Raises:
        HTTPException: If registration fails
    """
    return await auth_service.register_user(user_data)


@router.post(
    "/user-management/auth/login",
    tags=["authentication"],
    summary="Log in user",
    description="Authenticate user and return access token",
)
async def login() -> JSONResponse:
    """Log in user.

    Returns:
        JSONResponse: Login result with access token
    """
    # TODO: Implement user login
    return JSONResponse(content={"message": "Login endpoint"})


@router.post(
    "/user-management/auth/logout",
    tags=["authentication"],
    summary="Log out user",
    description="Invalidate user session",
)
async def logout() -> JSONResponse:
    """Log out user.

    Returns:
        JSONResponse: Logout confirmation
    """
    # TODO: Implement user logout
    return JSONResponse(content={"message": "Logout endpoint"})


@router.post(
    "/user-management/auth/refresh",
    tags=["authentication"],
    summary="Refresh access token",
    description="Get new access token using refresh token",
)
async def refresh_token() -> JSONResponse:
    """Refresh access token.

    Returns:
        JSONResponse: New access token
    """
    # TODO: Implement token refresh
    return JSONResponse(content={"message": "Token refresh endpoint"})


@router.post(
    "/user-management/auth/reset-password",
    tags=["authentication"],
    summary="Request password reset",
    description="Send password reset email",
)
async def request_password_reset() -> JSONResponse:
    """Request password reset.

    Returns:
        JSONResponse: Password reset request confirmation
    """
    # TODO: Implement password reset request
    return JSONResponse(content={"message": "Password reset request endpoint"})


@router.post(
    "/user-management/auth/reset-password/confirm",
    tags=["authentication"],
    summary="Confirm password reset",
    description="Reset password using token",
)
async def confirm_password_reset() -> JSONResponse:
    """Confirm password reset.

    Returns:
        JSONResponse: Password reset confirmation
    """
    # TODO: Implement password reset confirmation
    return JSONResponse(content={"message": "Password reset confirmation endpoint"})
