"""Notification management route handler.

Defines endpoints for notification management and preferences.
"""

from http import HTTPStatus
from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Path, Query, Response

from app.api.v1.schemas.request.notification.notification_delete_request import (
    NotificationDeleteRequest,
)
from app.api.v1.schemas.request.preference.update_user_preference_request import (
    UpdateUserPreferenceRequest,
)
from app.api.v1.schemas.response.error_response import ErrorResponse
from app.api.v1.schemas.response.notification.notification_count_response import (
    NotificationCountResponse,
)
from app.api.v1.schemas.response.notification.notification_delete_response import (
    NotificationDeleteResponse,
)
from app.api.v1.schemas.response.notification.notification_list_response import (
    NotificationListResponse,
)
from app.api.v1.schemas.response.notification.notification_read_all_response import (
    NotificationReadAllResponse,
)
from app.api.v1.schemas.response.notification.notification_read_response import (
    NotificationReadResponse,
)
from app.api.v1.schemas.response.preference.user_preference_response import (
    UserPreferenceResponse,
)
from app.db.sql.sql_database_manager import get_db
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.middleware.auth_middleware import get_current_user_id
from app.services.notification_service import NotificationService

router = APIRouter()


async def get_notification_service(
    db: Annotated[SqlDatabaseSession, Depends(get_db)],
) -> NotificationService:
    """Get notification service instance.

    Args:
        db: Database session

    Returns:
        NotificationService: Notification service instance
    """
    return NotificationService(db)


@router.get(
    "/user-management/notifications",
    tags=["notifications"],
    summary="Get notifications",
    description="Retrieve user's notifications",
    response_model=(NotificationListResponse | NotificationCountResponse),
    responses={
        HTTPStatus.OK: {
            "model": (NotificationListResponse | NotificationCountResponse),
            "description": "Notifications retrieved successfully",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Bad request",
        },
        HTTPStatus.UNAUTHORIZED: {
            "model": ErrorResponse,
            "description": "Invalid or missing authorization token",
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
async def get_notifications(
    authenticated_user_id: Annotated[str, Depends(get_current_user_id)],
    notification_service: Annotated[
        NotificationService,
        Depends(get_notification_service),
    ],
    limit: Annotated[
        int, Query(ge=1, le=100, description="Number of results to return")
    ] = 20,
    offset: Annotated[int, Query(ge=0, description="Number of results to skip")] = 0,
    count_only: Annotated[
        bool, Query(description="Return only the count of results")
    ] = False,
) -> NotificationListResponse | NotificationCountResponse:
    """Get notifications.

    Args:
        authenticated_user_id: User ID from JWT token
        limit: Number of results to return (1-100)
        offset: Number of results to skip
        count_only: Return only the count of results
        notification_service: Notification service instance

    Returns:
        NotificationListResponse | NotificationCountResponse:
            List of notifications or count
    """
    if offset > limit:
        raise HTTPException(
            status_code=HTTPStatus.BAD_REQUEST,
            detail="Offset cannot be greater than limit",
        )

    return await notification_service.get_notifications(
        user_id=UUID(authenticated_user_id),
        limit=limit,
        offset=offset,
        count_only=count_only,
    )


@router.put(
    "/user-management/notifications/{notification_id}/read",
    tags=["notifications"],
    summary="Mark notification as read",
    description="Mark a notification as read",
    response_model=NotificationReadResponse,
    responses={
        HTTPStatus.OK: {
            "model": NotificationReadResponse,
            "description": "Notification marked as read successfully",
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
            "description": "Notification not found",
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
async def mark_notification_read(
    notification_id: Annotated[UUID, Path(description="Notification ID")],
    authenticated_user_id: Annotated[str, Depends(get_current_user_id)],
    notification_service: Annotated[
        NotificationService,
        Depends(get_notification_service),
    ],
) -> NotificationReadResponse:
    """Mark notification as read.

    Args:
        notification_id: The notification's unique identifier
        authenticated_user_id: User ID from JWT token
        notification_service: Notification service instance

    Returns:
        NotificationReadResponse: Read confirmation
    """
    return await notification_service.mark_notification_read(
        UUID(authenticated_user_id), notification_id
    )


@router.put(
    "/user-management/notifications/read-all",
    tags=["notifications"],
    summary="Mark all notifications as read",
    description="Mark all notifications as read",
    response_model=NotificationReadAllResponse,
    responses={
        HTTPStatus.OK: {
            "model": NotificationReadAllResponse,
            "description": "All notifications marked as read successfully",
        },
        HTTPStatus.BAD_REQUEST: {
            "model": ErrorResponse,
            "description": "Bad request",
        },
        HTTPStatus.UNAUTHORIZED: {
            "model": ErrorResponse,
            "description": "Invalid or missing authorization token",
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
async def mark_all_notifications_read(
    authenticated_user_id: Annotated[str, Depends(get_current_user_id)],
    notification_service: Annotated[
        NotificationService,
        Depends(get_notification_service),
    ],
) -> NotificationReadAllResponse:
    """Mark all notifications as read.

    Args:
        authenticated_user_id: User ID from JWT token
        notification_service: Notification service instance

    Returns:
        NotificationReadAllResponse: Read all confirmation
    """
    return await notification_service.mark_all_notifications_read(
        UUID(authenticated_user_id)
    )


@router.delete(
    "/user-management/notifications",
    tags=["notifications"],
    summary="Delete notifications",
    description="Delete multiple notifications",
    response_model=NotificationDeleteResponse,
    status_code=HTTPStatus.OK,
    responses={
        HTTPStatus.OK: {
            "model": NotificationDeleteResponse,
            "description": "All notifications deleted successfully",
        },
        HTTPStatus.PARTIAL_CONTENT: {
            "model": NotificationDeleteResponse,
            "description": (
                "Partial success - some notifications deleted, others not found"
            ),
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
            "description": "No notifications found",
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
async def delete_notifications(
    request: NotificationDeleteRequest,
    authenticated_user_id: Annotated[str, Depends(get_current_user_id)],
    notification_service: Annotated[
        NotificationService,
        Depends(get_notification_service),
    ],
) -> NotificationDeleteResponse | Response:
    """Delete notifications.

    Args:
        request: Request containing notification IDs to delete
        authenticated_user_id: User ID from JWT token
        notification_service: Notification service instance

    Returns:
        NotificationDeleteResponse: Deletion confirmation
    """
    response, is_partial_success = await notification_service.delete_notifications(
        UUID(authenticated_user_id), request.notification_ids
    )

    if is_partial_success:
        return Response(
            content=response.model_dump_json(),
            status_code=HTTPStatus.PARTIAL_CONTENT,
            media_type="application/json",
        )

    return response


@router.get(
    "/user-management/notifications/preferences",
    tags=["notifications"],
    summary="Get notification preferences",
    description="Retrieve user's notification preferences",
    response_model=UserPreferenceResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserPreferenceResponse,
            "description": "Notification preferences retrieved successfully",
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
async def get_notification_preferences(
    authenticated_user_id: Annotated[str, Depends(get_current_user_id)],
    notification_service: Annotated[
        NotificationService,
        Depends(get_notification_service),
    ],
) -> UserPreferenceResponse:
    """Get notification preferences.

    Returns:
        NotificationPreferencesResponse: Notification preferences
    """
    return await notification_service.get_notification_preferences(
        user_id=UUID(authenticated_user_id)
    )


@router.put(
    "/user-management/notifications/preferences",
    tags=["notifications"],
    summary="Update notification preferences",
    description="Update user's notification preferences",
    response_model=UserPreferenceResponse,
    responses={
        HTTPStatus.OK: {
            "model": UserPreferenceResponse,
            "description": "Notification preferences updated successfully",
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
async def update_notification_preferences(
    preferences: UpdateUserPreferenceRequest,
    authenticated_user_id: Annotated[str, Depends(get_current_user_id)],
    notification_service: Annotated[
        NotificationService,
        Depends(get_notification_service),
    ],
) -> UserPreferenceResponse:
    """Update notification preferences.

    Args:
        preferences: UpdateUserPreferenceRequest body
        authenticated_user_id: User ID from JWT token
        notification_service: Notification service instance

    Returns:
        UserPreferenceResponse: Updated preferences
    """
    return await notification_service.update_user_preferences(
        UUID(authenticated_user_id), preferences
    )
