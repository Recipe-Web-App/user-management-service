"""Notification management route handler.

Defines endpoints for notification management and preferences.
"""

from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, Path, Query
from fastapi.responses import JSONResponse

from app.api.v1.schemas.response.notification_count_response import (
    NotificationCountResponse,
)
from app.api.v1.schemas.response.notification_list_response import (
    NotificationListResponse,
)
from app.api.v1.schemas.response.notification_read_all_response import (
    NotificationReadAllResponse,
)
from app.api.v1.schemas.response.notification_read_response import (
    NotificationReadResponse,
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
    "/user-management/users/notifications/{notification_id}",
    tags=["notifications"],
    summary="Delete notification",
    description="Delete a specific notification",
)
async def delete_notification(
    notification_id: Annotated[UUID, Path(description="Notification ID")],
) -> JSONResponse:
    """Delete notification.

    Args:
        user_id: The user's unique identifier
        notification_id: The notification's unique identifier

    Returns:
        JSONResponse: Deletion confirmation
    """
    # TODO: Implement delete notification
    return JSONResponse(content={"message": f"Delete notification {notification_id}"})


@router.get(
    "/user-management/users/{user_id}/notifications/preferences",
    tags=["notifications"],
    summary="Get notification preferences",
    description="Retrieve user's notification preferences",
)
async def get_notification_preferences(
    user_id: Annotated[UUID, Path(description="User ID")],
) -> JSONResponse:
    """Get notification preferences.

    Returns:
        JSONResponse: Notification preferences
    """
    # TODO: Implement get notification preferences
    return JSONResponse(content={"message": f"Get {user_id} preferences endpoint"})


@router.put(
    "/user-management/users/{user_id}/notifications/preferences",
    tags=["notifications"],
    summary="Update notification preferences",
    description="Update user's notification preferences",
)
async def update_notification_preferences(
    user_id: Annotated[UUID, Path(description="User ID")],
) -> JSONResponse:
    """Update notification preferences.

    Returns:
        JSONResponse: Updated preferences
    """
    # TODO: Implement update notification preferences
    return JSONResponse(content={"message": f"Update {user_id} preferences endpoint"})
