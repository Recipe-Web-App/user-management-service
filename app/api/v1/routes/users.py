"""User management route handler.

Defines endpoints for user profile management and account operations.
"""

from http import HTTPStatus
from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Path, Query

from app.api.v1.schemas.request.user.user_account_delete_request import (
    UserAccountDeleteRequest,
)
from app.api.v1.schemas.request.user.user_confirm_account_delete_response import (
    UserConfirmAccountDeleteResponse,
)
from app.api.v1.schemas.request.user.user_profile_update_request import (
    UserProfileUpdateRequest,
)
from app.api.v1.schemas.response.error_response import ErrorResponse
from app.api.v1.schemas.response.user.user_account_delete_response import (
    UserAccountDeleteRequestResponse,
)
from app.api.v1.schemas.response.user.user_profile_response import UserProfileResponse
from app.api.v1.schemas.response.user.user_search_response import (
    UserSearchResponse,
    UserSearchResult,
)
from app.deps.auth import ReadScopeDep, WriteScopeDep
from app.deps.services import UserServiceDep
from app.middleware.auth_middleware import get_current_user_id, get_optional_user_id

router = APIRouter()


@router.get(
    "/user-management/users/{user_id}/profile",
    tags=["users"],
    summary="Get user profile",
    description=(
        "Retrieve user profile information with privacy checks. "
        "**Required OAuth2 Scope**: `user:read`"
    ),
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
    user_context: ReadScopeDep,
    user_service: UserServiceDep,
) -> UserProfileResponse:
    """Get user profile.

    Args:
        user_id: The user's unique identifier
        user_context: OAuth2 user context with required user:read scope
        user_service: User service instance

    Returns:
        UserProfileResponse: User profile data with optional preferences

    Raises:
        HTTPException: If user not found, forbidden, or database error
    """
    return await user_service.get_user_profile(
        user_id=user_id,
        requester_user_id=UUID(user_context.user_id),
    )


@router.put(
    "/user-management/users/profile",
    tags=["users"],
    summary="Update user profile",
    description=(
        "Update current user's profile information. "
        "**Required OAuth2 Scope**: `user:write`"
    ),
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
    user_context: WriteScopeDep,
    user_service: UserServiceDep,
) -> UserProfileResponse:
    """Update user profile.

    Args:
        update_data: The profile data to update
        user_context: OAuth2 user context with required user:write scope
        user_service: User service instance

    Returns:
        UserProfileResponse: Updated profile data

    Raises:
        HTTPException: If user not found, validation error, or database error
    """
    return await user_service.update_user_profile(
        user_id=UUID(user_context.user_id),
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
    user_service: UserServiceDep,
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
    response_model=UserConfirmAccountDeleteResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserConfirmAccountDeleteResponse,
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
    user_service: UserServiceDep,
) -> UserConfirmAccountDeleteResponse:
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
    response_model=UserSearchResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserSearchResponse,
            "description": "User search results retrieved successfully",
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
async def search_users(  # noqa: PLR0913
    authenticated_user_id: Annotated[str, Depends(get_current_user_id)],
    user_service: UserServiceDep,
    query: Annotated[
        str | None, Query(description="Search query for username or display name")
    ] = None,
    limit: Annotated[
        int, Query(ge=1, le=100, description="Number of results to return")
    ] = 20,
    offset: Annotated[int, Query(ge=0, description="Number of results to skip")] = 0,
    count_only: Annotated[
        bool, Query(description="Return only the count of results")
    ] = False,
) -> UserSearchResponse:
    """Search users with privacy checks.

    Args:
        authenticated_user_id: The authenticated user making the request
        user_service: User service instance
        query: Search query for username or display name
        limit: Number of results to return (1-100)
        offset: Number of results to skip
        count_only: Return only the count of results

    Returns:
        UserSearchResponse: Paginated user search results
    """
    if offset > limit:
        raise HTTPException(
            status_code=HTTPStatus.BAD_REQUEST,
            detail="Offset cannot be greater than limit.",
        )
    return await user_service.search_users(
        requester_user_id=UUID(authenticated_user_id),
        query=query,
        limit=limit,
        offset=offset,
        count_only=count_only,
    )


@router.get(
    "/user-management/users/{user_id}",
    tags=["users"],
    summary="Get user by ID",
    description=(
        "Retrieve public profile of another user. Respects privacy settings - "
        "private profiles may not be accessible to anonymous users or other users "
        "depending on their privacy preferences."
    ),
    response_model=UserSearchResult,
    responses={
        HTTPStatus.OK: {
            "model": UserSearchResult,
            "description": "Public user profile retrieved successfully",
        },
        HTTPStatus.NOT_FOUND: {
            "model": ErrorResponse,
            "description": "User not found or not accessible due to privacy settings",
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
async def get_user_by_id(
    user_id: Annotated[UUID, Path(description="User ID")],
    requester_user_id: Annotated[str | None, Depends(get_optional_user_id)],
    user_service: UserServiceDep,
) -> UserSearchResult:
    """Get public user profile by ID.

    Args:
        user_id: The user's unique identifier
        requester_user_id: The authenticated user making the request (optional)
        user_service: User service instance

    Returns:
        UserSearchResult: Public user profile data

    Raises:
        HTTPException: If user not found or access denied due to privacy settings
    """
    requester_uuid = UUID(requester_user_id) if requester_user_id else None
    return await user_service.get_public_user_by_id(
        user_id=user_id,
        requester_user_id=requester_uuid,
    )
