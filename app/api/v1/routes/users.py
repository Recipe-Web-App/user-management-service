"""User management route handler.

Defines endpoints for user profile management and account operations.
"""

from http import HTTPStatus
from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, Path, Query
from fastapi.responses import JSONResponse

from app.api.v1.schemas.request.user.user_account_delete_request import (
    UserAccountDeleteRequest,
)
from app.api.v1.schemas.request.user.user_profile_update_request import (
    UserProfileUpdateRequest,
)
from app.api.v1.schemas.response.error_response import ErrorResponse
from app.api.v1.schemas.response.user.user_account_delete_response import (
    UserAccountDeleteRequestResponse,
)
from app.api.v1.schemas.response.user.user_profile_response import UserProfileResponse
from app.db.redis.redis_database_manager import get_redis_session
from app.db.redis.redis_database_session import RedisDatabaseSession
from app.db.sql.sql_database_manager import get_db
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.middleware.auth_middleware import get_current_user_id
from app.services.user_service import UserService

router = APIRouter()


async def get_user_service(
    db: Annotated[SqlDatabaseSession, Depends(get_db)],
    redis_session: Annotated[RedisDatabaseSession, Depends(get_redis_session)],
) -> UserService:
    """Get user service instance with SQL and Redis sessions."""
    return UserService(db, redis_session)


@router.get(
    "/user-management/users/{user_id}/profile",
    tags=["users"],
    summary="Get user profile",
    description="Retrieve user profile information with privacy checks",
    response_model=UserProfileResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserProfileResponse,
            "description": "User profile retrieved successfully",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Bad request",
        },
        HTTPStatus.UNAUTHORIZED: {
            "model": ErrorResponse,
            "description": "Invalid or missing authorization token",
        },
        HTTPStatus.FORBIDDEN: {
            "model": ErrorResponse,
            "description": "Profile access forbidden due to privacy settings",
        },
        HTTPStatus.NOT_FOUND: {
            "model": ErrorResponse,
            "description": "User not found",
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
async def get_profile(
    user_id: Annotated[UUID, Path(description="User ID")],
    authenticated_user_id: Annotated[str, Depends(get_current_user_id)],
    user_service: Annotated[UserService, Depends(get_user_service)],
) -> UserProfileResponse:
    """Get user profile.

    Args:
        user_id: The user's unique identifier
        authenticated_user_id: The authenticated user making the request
        user_service: User service instance

    Returns:
        UserProfileResponse: User profile data with optional preferences

    Raises:
        HTTPException: If user not found, forbidden, or database error
    """
    return await user_service.get_user_profile(
        user_id=user_id,
        requester_user_id=UUID(authenticated_user_id),
    )


@router.put(
    "/user-management/users/profile",
    tags=["users"],
    summary="Update user profile",
    description="Update current user's profile information",
    response_model=UserProfileResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserProfileResponse,
            "description": "Profile updated successfully",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Bad request (e.g., username/email already exists)",
        },
        HTTPStatus.UNAUTHORIZED: {
            "model": ErrorResponse,
            "description": "Invalid or missing authorization token",
        },
        HTTPStatus.NOT_FOUND: {
            "model": ErrorResponse,
            "description": "User not found",
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
async def update_profile(
    update_data: UserProfileUpdateRequest,
    authenticated_user_id: Annotated[str, Depends(get_current_user_id)],
    user_service: Annotated[UserService, Depends(get_user_service)],
) -> UserProfileResponse:
    """Update user profile.

    Args:
        update_data: The profile data to update
        authenticated_user_id: The authenticated user making the request
        user_service: User service instance

    Returns:
        UserProfileResponse: Updated profile data

    Raises:
        HTTPException: If user not found, validation error, or database error
    """
    return await user_service.update_user_profile(
        user_id=UUID(authenticated_user_id),
        update_data=update_data,
    )


@router.post(
    "/user-management/users/account/delete-request",
    tags=["users"],
    summary="Request account deletion",
    description="Request account deletion and receive a confirmation token",
    response_model=UserAccountDeleteRequestResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserAccountDeleteRequestResponse,
            "description": "Deletion request created successfully",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Account already inactive",
        },
        HTTPStatus.UNAUTHORIZED: {
            "model": ErrorResponse,
            "description": "Invalid or missing authorization token",
        },
        HTTPStatus.NOT_FOUND: {
            "model": ErrorResponse,
            "description": "User not found",
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
async def request_account_deletion(
    authenticated_user_id: Annotated[str, Depends(get_current_user_id)],
    user_service: Annotated[UserService, Depends(get_user_service)],
) -> UserAccountDeleteRequestResponse:
    """Request account deletion.

    Args:
        authenticated_user_id: The authenticated user making the request
        user_service: User service instance

    Returns:
        UserAccountDeleteRequestResponse: Deletion request with confirmation token

    Raises:
        HTTPException: If user not found or database error
    """
    return await user_service.request_account_deletion(
        user_id=UUID(authenticated_user_id),
    )


@router.delete(
    "/user-management/users/account",
    tags=["users"],
    summary="Confirm account deletion",
    description="Confirm account deletion using the confirmation token",
    response_model=dict,
    responses={
        HTTPStatus.OK: {
            "model": dict,
            "description": "Account successfully deactivated",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Invalid or expired confirmation token",
        },
        HTTPStatus.UNAUTHORIZED: {
            "model": ErrorResponse,
            "description": "Invalid or missing authorization token",
        },
        HTTPStatus.NOT_FOUND: {
            "model": ErrorResponse,
            "description": "User not found",
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
async def confirm_account_deletion(
    delete_request: UserAccountDeleteRequest,
    authenticated_user_id: Annotated[str, Depends(get_current_user_id)],
    user_service: Annotated[UserService, Depends(get_user_service)],
) -> dict:
    """Confirm account deletion.

    Args:
        delete_request: The deletion confirmation request
        authenticated_user_id: The authenticated user making the request
        user_service: User service instance

    Returns:
        dict: Confirmation of account deactivation

    Raises:
        HTTPException: If user not found, invalid token, or database error
    """
    return await user_service.confirm_account_deletion(
        user_id=UUID(authenticated_user_id),
        delete_request=delete_request,
    )


@router.get(
    "/user-management/users/search",
    tags=["users"],
    summary="Search users",
    description="Search for users by username or display name",
)
async def search_users(
    limit: Annotated[
        int, Query(ge=1, le=100, description="Number of results to return")
    ] = 20,
    offset: Annotated[int, Query(ge=0, description="Number of results to skip")] = 0,
    count_only: Annotated[
        bool, Query(description="Return only the count of results")
    ] = False,
) -> JSONResponse:
    """Search users.

    Args:
        limit: Number of results to return (1-100)
        offset: Number of results to skip
        count_only: Return only the count of results

    Returns:
        JSONResponse: Search results or count
    """
    # TODO: Implement user search
    return JSONResponse(
        content={
            "message": "Search users endpoint",
            "pagination": {
                "limit": limit,
                "offset": offset,
                "count_only": count_only,
            },
        }
    )


@router.get(
    "/user-management/users/{user_id}",
    tags=["users"],
    summary="Get user by ID",
    description="Retrieve public profile of another user",
)
async def get_user_by_id(
    user_id: Annotated[UUID, Path(description="User ID")],
) -> JSONResponse:
    """Get user by ID.

    Args:
        user_id: The user's unique identifier

    Returns:
        JSONResponse: Public user profile data
    """
    # TODO: Implement get user by ID
    return JSONResponse(content={"message": f"Get user {user_id} endpoint"})
