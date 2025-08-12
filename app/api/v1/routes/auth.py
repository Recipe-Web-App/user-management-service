"""Authentication routes.

Provides endpoints for user authentication including registration, login, logout, token
refresh, and password reset.
"""

from http import HTTPStatus
from typing import Annotated

from fastapi import APIRouter, Body, Depends

from app.api.v1.schemas.request.user.user_login_request import UserLoginRequest
from app.api.v1.schemas.request.user.user_password_reset_confirm_request import (
    UserPasswordResetConfirmRequest,
)
from app.api.v1.schemas.request.user.user_password_reset_request import (
    UserPasswordResetRequest,
)
from app.api.v1.schemas.request.user.user_refresh_request import UserRefreshRequest
from app.api.v1.schemas.request.user.user_registration_request import (
    UserRegistrationRequest,
)
from app.api.v1.schemas.response.error_response import ErrorResponse
from app.api.v1.schemas.response.user.user_login_response import UserLoginResponse
from app.api.v1.schemas.response.user.user_logout_response import UserLogoutResponse
from app.api.v1.schemas.response.user.user_password_reset_confirm_response import (
    UserPasswordResetConfirmResponse,
)
from app.api.v1.schemas.response.user.user_password_reset_response import (
    UserPasswordResetResponse,
)
from app.api.v1.schemas.response.user.user_refresh_response import UserRefreshResponse
from app.api.v1.schemas.response.user.user_registration_response import (
    UserRegistrationResponse,
)
from app.deps.services import AuthServiceDep
from app.middleware.auth_middleware import get_current_user_id

router = APIRouter()


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
    auth_service: AuthServiceDep,
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
    auth_service: AuthServiceDep,
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
    auth_service: AuthServiceDep,
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
    response_model=UserRefreshResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserRefreshResponse,
            "description": "Token refreshed successfully",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Bad request",
        },
        HTTPStatus.UNAUTHORIZED: {
            "model": ErrorResponse,
            "description": "Invalid refresh token or no active session",
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
async def refresh_token(
    refresh_data: Annotated[UserRefreshRequest, Body(...)],
    auth_service: AuthServiceDep,
) -> UserRefreshResponse:
    """Refresh access token.

    Args:
        refresh_data: Refresh token data
        auth_service: Auth service instance

    Returns:
        UserRefreshResponse: New access token

    Raises:
        HTTPException: If refresh token is invalid, expired, or user not found
    """
    return await auth_service.refresh_token(refresh_data)


@router.post(
    "/user-management/auth/reset-password",
    tags=["authentication"],
    summary="Request password reset",
    description="Send password reset email",
    response_model=UserPasswordResetResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserPasswordResetResponse,
            "description": "Password reset email sent successfully",
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
async def request_password_reset(
    reset_data: Annotated[UserPasswordResetRequest, Body(...)],
    auth_service: AuthServiceDep,
) -> UserPasswordResetResponse:
    """Request password reset.

    Args:
        reset_data: Password reset request data
        auth_service: Auth service instance

    Returns:
        UserPasswordResetResponse: Password reset request result

    Raises:
        HTTPException: If password reset request fails
    """
    return await auth_service.request_password_reset(reset_data)


@router.post(
    "/user-management/auth/reset-password/confirm",
    tags=["authentication"],
    summary="Confirm password reset",
    description="Reset password using token",
    response_model=UserPasswordResetConfirmResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserPasswordResetConfirmResponse,
            "description": "Password reset successfully",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Invalid or expired reset token",
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
async def confirm_password_reset(
    confirm_data: Annotated[UserPasswordResetConfirmRequest, Body(...)],
    auth_service: AuthServiceDep,
) -> UserPasswordResetConfirmResponse:
    """Confirm password reset.

    Args:
        confirm_data: Password reset confirmation data
        auth_service: Auth service instance

    Returns:
        UserPasswordResetConfirmResponse: Password reset confirmation result

    Raises:
        HTTPException: If password reset confirmation fails
    """
    return await auth_service.confirm_password_reset(confirm_data)
