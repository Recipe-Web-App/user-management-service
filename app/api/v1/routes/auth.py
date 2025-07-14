"""Authentication route handler.

Defines endpoints for user authentication and account management.
"""

from http import HTTPStatus
from typing import Annotated

from fastapi import APIRouter, Body, Depends
from fastapi.responses import JSONResponse

from app.api.v1.schemas.request.user_login_request import UserLoginRequest
from app.api.v1.schemas.request.user_registration_request import UserRegistrationRequest
from app.api.v1.schemas.response.error_response import ErrorResponse
from app.api.v1.schemas.response.user_login_response import UserLoginResponse
from app.api.v1.schemas.response.user_logout_response import UserLogoutResponse
from app.api.v1.schemas.response.user_registration_response import (
    UserRegistrationResponse,
)
from app.db.redis.redis_database_manager import get_redis_session
from app.db.redis.redis_database_session import RedisDatabaseSession
from app.db.sql.sql_database_manager import get_db
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.middleware import get_current_user_id
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
        HTTPStatus.SERVICE_UNAVAILABLE: {
            "model": ErrorResponse,
            "description": "Service temporarily unavailable",
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
    response_model=UserLoginResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserLoginResponse,
            "description": "User logged in successfully",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Bad request",
        },
        HTTPStatus.UNAUTHORIZED: {
            "model": ErrorResponse,
            "description": "Invalid credentials or inactive account",
        },
        HTTPStatus.UNPROCESSABLE_ENTITY: {
            "model": ErrorResponse,
            "description": "Validation error",
        },
        HTTPStatus.INTERNAL_SERVER_ERROR: {
            "model": ErrorResponse,
            "description": "Internal server error",
        },
        HTTPStatus.SERVICE_UNAVAILABLE: {
            "model": ErrorResponse,
            "description": "Service temporarily unavailable",
        },
        HTTPStatus.CONFLICT: {
            "model": ErrorResponse,
            "description": "User already logged in",
        },
    },
)
async def login(
    login_data: Annotated[UserLoginRequest, Body(...)],
    auth_service: Annotated[AuthService, Depends(get_auth_service)],
) -> UserLoginResponse:
    """Log in user.

    Args:
        login_data: User login credentials
        auth_service: Auth service instance

    Returns:
        UserLoginResponse: Login result with user data and access token

    Raises:
        HTTPException: If login fails due to invalid credentials, inactive account,
        or service unavailability
    """
    return await auth_service.login_user(login_data)


@router.post(
    "/user-management/auth/logout",
    tags=["authentication"],
    summary="Log out user",
    description="Invalidate user session",
    response_model=UserLogoutResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserLogoutResponse,
            "description": "User logged out successfully",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Bad request",
        },
        HTTPStatus.UNAUTHORIZED: {
            "model": ErrorResponse,
            "description": "Invalid or missing authorization token",
        },
        HTTPStatus.INTERNAL_SERVER_ERROR: {
            "model": ErrorResponse,
            "description": "Internal server error",
        },
        HTTPStatus.SERVICE_UNAVAILABLE: {
            "model": ErrorResponse,
            "description": "Service temporarily unavailable",
        },
    },
)
async def logout(
    user_id: Annotated[str, Depends(get_current_user_id)],
    auth_service: Annotated[AuthService, Depends(get_auth_service)],
) -> UserLogoutResponse:
    """Log out user.

    Args:
        user_id: User ID extracted from JWT token
        auth_service: Auth service instance

    Returns:
        UserLogoutResponse: Logout result with confirmation

    Raises:
        HTTPException: If logout fails due to invalid token or service unavailability
    """
    return await auth_service.logout_user(user_id)


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
