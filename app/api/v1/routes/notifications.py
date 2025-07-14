"""Notification management route handler.

Defines endpoints for notification management and preferences.
"""

from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Path, Query, status
from fastapi.responses import JSONResponse

from app.api.v1.schemas.response.notification_response import (
    NotificationCountResponse,
    NotificationListResponse,
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
    "/user-management/users/{user_id}/notifications",
    tags=["notifications"],
    summary="Get notifications",
    description="Retrieve user's notifications",
    response_model=(NotificationListResponse | NotificationCountResponse),
)
async def get_notifications(  # noqa: PLR0913
    user_id: Annotated[UUID, Path(description="User ID")],
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
        user_id: The user's unique identifier
        authenticated_user_id: User ID from JWT token
        limit: Number of results to return (1-100)
        offset: Number of results to skip
        count_only: Return only the count of results
        notification_service: Notification service instance

    Returns:
        NotificationListResponse | NotificationCountResponse:
            List of notifications or count
    """
    # Verify user can only access their own notifications
    if str(user_id) != authenticated_user_id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="You can only access your own notifications",
        )

    return await notification_service.get_notifications(
        user_id=user_id,
        limit=limit,
        offset=offset,
        count_only=count_only,
    )


@router.put(
    "/user-management/users//notifications/{notification_id}/read",
    tags=["notifications"],
    summary="Mark notification as read",
    description="Mark a notification as read",
)
async def mark_notification_read(
    notification_id: Annotated[UUID, Path(description="Notification ID")],
) -> JSONResponse:
    """Mark notification as read.

    Args:
        user_id: The user's unique identifier
        notification_id: The notification's unique identifier

    Returns:
        JSONResponse: Read confirmation
    """
    # TODO: Implement mark notification read
    return JSONResponse(
        content={"message": f"Mark notification {notification_id} as read for user"}
    )


@router.put(
    "/user-management/users/{user_id}/notifications/read-all",
    tags=["notifications"],
    summary="Mark all notifications as read",
    description="Mark all notifications as read",
)
async def mark_all_notifications_read(
    user_id: Annotated[UUID, Path(description="User ID")],
) -> JSONResponse:
    """Mark all notifications as read.

    Returns:
        JSONResponse: Read all confirmation
    """
    # TODO: Implement mark all notifications read
    return JSONResponse(
        content={"message": f"Mark all {user_id} notifications as read"}
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
