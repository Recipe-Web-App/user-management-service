"""Social features route handler.

Defines endpoints for social interactions like following, followers, and activity.
"""

from http import HTTPStatus
from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Path, Query

from app.api.v1.schemas.response.error_response import ErrorResponse
from app.api.v1.schemas.response.social import GetFollowedUsersResponse
from app.db.sql.sql_database_manager import get_db
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.middleware.auth_middleware import get_current_user_id
from app.services.social_service import SocialService

router = APIRouter()


async def get_social_service(
    db: Annotated[SqlDatabaseSession, Depends(get_db)],
) -> SocialService:
    """Get social service instance.

    Args:
        db: Database session

    Returns:
        SocialService: Social service instance
    """
    return SocialService(db)


@router.get(
    "/user-management/following",
    tags=["social"],
    summary="Get following list",
    description="Retrieve list of users the current user is following",
    response_model=GetFollowedUsersResponse,
    responses={
        HTTPStatus.OK: {
            "model": GetFollowedUsersResponse,
            "description": "Following list retrieved successfully",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Bad request",
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
async def get_followed_users(
    authenticated_user_id: Annotated[str, Depends(get_current_user_id)],
    social_service: Annotated[
        SocialService,
        Depends(get_social_service),
    ],
    limit: Annotated[
        int, Query(ge=1, le=100, description="Number of results to return")
    ] = 20,
    offset: Annotated[int, Query(ge=0, description="Number of results to skip")] = 0,
    count_only: Annotated[
        bool, Query(description="Return only the count of results")
    ] = False,
) -> GetFollowedUsersResponse:
    """Get following list.

    Args:
        authenticated_user_id: User ID from JWT token
        limit: Number of results to return (1-100)
        offset: Number of results to skip
        count_only: Return only the count of results
        social_service: Social service instance

    Returns:
        GetFollowedUsersResponse: List of users being followed or count
    """
    if offset > limit:
        raise HTTPException(
            status_code=HTTPStatus.BAD_REQUEST,
            detail="Offset cannot be greater than limit.",
        )
    return await social_service.get_followed_users(
        user_id=UUID(authenticated_user_id),
        limit=limit,
        offset=offset,
        count_only=count_only,
    )


@router.get(
    "/user-management/users/{user_id}/followers",
    tags=["social"],
    summary="Get followers list",
    description="Retrieve list of users following the current user",
    response_model=GetFollowedUsersResponse,
    responses={
        HTTPStatus.OK: {
            "model": GetFollowedUsersResponse,
            "description": "Followers list retrieved successfully",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Bad request",
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
async def get_followers(
    user_id: Annotated[UUID, Path(description="User ID")],
    social_service: Annotated[
        SocialService,
        Depends(get_social_service),
    ],
    limit: Annotated[
        int, Query(ge=1, le=100, description="Number of results to return")
    ] = 20,
    offset: Annotated[int, Query(ge=0, description="Number of results to skip")] = 0,
    count_only: Annotated[
        bool, Query(description="Return only the count of results")
    ] = False,
) -> GetFollowedUsersResponse:
    """Get followers list.

    Args:
        user_id: The user's unique identifier
        limit: Number of results to return (1-100)
        offset: Number of results to skip
        count_only: Return only the count of results
        social_service: Social service instance

    Returns:
        GetFollowedUsersResponse: List of followers or count
    """
    if offset > limit:
        raise HTTPException(
            status_code=HTTPStatus.BAD_REQUEST,
            detail="Offset cannot be greater than limit.",
        )
    return await social_service.get_followers(
        user_id=user_id,
        limit=limit,
        offset=offset,
        count_only=count_only,
    )


@router.post(
    "/user-management/users/{user_id}/follow/{target_user_id}",
    tags=["social"],
    summary="Follow user",
    description="Follow another user",
)
async def follow_user(
    user_id: Annotated[UUID, Path(description="User ID")],
    target_user_id: Annotated[UUID, Path(description="Target user ID to follow")],
) -> dict:
    """Follow user.

    Args:
        user_id: The user's unique identifier
        target_user_id: The user's unique identifier to follow

    Returns:
        dict: Follow confirmation
    """
    # TODO: Implement follow user
    return {"message": f"User {user_id} follow {target_user_id}"}


@router.delete(
    "/user-management/users/{user_id}/follow/{target_user_id}",
    tags=["social"],
    summary="Unfollow user",
    description="Unfollow another user",
)
async def unfollow_user(
    user_id: Annotated[UUID, Path(description="User ID")],
    target_user_id: Annotated[UUID, Path(description="Target user ID to unfollow")],
) -> dict:
    """Unfollow user.

    Args:
        user_id: The user's unique identifier
        target_user_id: The user's unique identifier to unfollow

    Returns:
        dict: Unfollow confirmation
    """
    # TODO: Implement unfollow user
    return {"message": f"User {user_id} unfollow {target_user_id}"}


@router.get(
    "/user-management/users/{user_id}/activity",
    tags=["social"],
    summary="Get user activity",
    description="Retrieve activity from a specific user",
)
async def get_user_activity(
    user_id: Annotated[UUID, Path(description="User ID")],
    limit: Annotated[
        int, Query(ge=1, le=100, description="Number of results to return")
    ] = 20,
    offset: Annotated[int, Query(ge=0, description="Number of results to skip")] = 0,
    count_only: Annotated[
        bool, Query(description="Return only the count of results")
    ] = False,
) -> dict:
    """Get user activity.

    Args:
        user_id: The user's unique identifier
        limit: Number of results to return (1-100)
        offset: Number of results to skip
        count_only: Return only the count of results

    Returns:
        dict: User activity data or count
    """
    # TODO: Implement get user activity
    return {
        "message": f"Get activity for user {user_id}",
        "pagination": {
            "limit": limit,
            "offset": offset,
            "count_only": count_only,
        },
    }
