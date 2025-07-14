"""Notification service for user management."""

from uuid import UUID

from fastapi import HTTPException, status
from sqlalchemy import func, select
from sqlalchemy.exc import DisconnectionError
from sqlalchemy.exc import TimeoutError as SQLTimeoutError

from app.api.v1.schemas.common.notification import Notification as NotificationSchema
from app.api.v1.schemas.response.notification_count_response import (
    NotificationCountResponse,
)
from app.api.v1.schemas.response.notification_list_response import (
    NotificationListResponse,
)
from app.api.v1.schemas.response.notification_read_response import (
    NotificationReadResponse,
)
from app.core.logging import get_logger
from app.db.sql.models.user.notification import Notification as NotificationModel
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.exceptions.custom_exceptions.database_exceptions import DatabaseError

_log = get_logger(__name__)


class NotificationService:
    """Service for notification operations."""

    def __init__(self, db: SqlDatabaseSession) -> None:
        """Initialize notification service with database session."""
        self.db = db

    async def get_notifications(
        self, user_id: UUID, limit: int = 20, offset: int = 0, count_only: bool = False
    ) -> NotificationListResponse | NotificationCountResponse:
        """Get notifications for a user.

        Args:
            user_id: The user's unique identifier
            limit: Number of results to return (1-100)
            offset: Number of results to skip
            count_only: Return only the count of results

        Returns:
            NotificationListResponse | NotificationCountResponse: Notifications or count

        Raises:
            HTTPException: If user not found or database services are unavailable
        """
        _log.info(f"Getting notifications for user: {user_id}")

        try:
            # Verify user exists
            user = await self.db.get_user_by_id(str(user_id))
            if not user:
                _log.warning(f"User not found: {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="User not found",
                )

            if count_only:
                # Get only count
                count_result = await self.db.execute(
                    select(func.count(NotificationModel.notification_id)).where(
                        NotificationModel.user_id == user_id,
                        NotificationModel.is_deleted.is_(False),
                    )
                )
                total_count = count_result.scalar()
                _log.info(f"Notification count for user {user_id}: {total_count}")
                return NotificationCountResponse(total_count=total_count)

            # Get notifications with pagination
            notifications_result = await self.db.execute(
                select(NotificationModel)
                .where(
                    NotificationModel.user_id == user_id,
                    NotificationModel.is_deleted.is_(False),
                )
                .order_by(NotificationModel.created_at.desc())
                .limit(limit)
                .offset(offset)
            )
            notifications = notifications_result.scalars().all()

            # Get total count for pagination
            count_result = await self.db.execute(
                select(func.count(NotificationModel.notification_id)).where(
                    NotificationModel.user_id == user_id,
                    NotificationModel.is_deleted.is_(False),
                )
            )
            total_count = count_result.scalar()

            # Convert to response schemas
            notification_schemas = [
                NotificationSchema.model_validate(notification)
                for notification in notifications
            ]

            _log.info(
                f"Retrieved {len(notification_schemas)} notifications "
                f"for user {user_id}"
            )

            return NotificationListResponse(
                notifications=notification_schemas,
                total_count=total_count,
                limit=limit,
                offset=offset,
            )

        except DatabaseError as e:
            _log.error(f"Database error while getting notifications: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            _log.error(f"Database connection error while getting notifications: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Database service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e

    async def mark_notification_read(
        self, user_id: UUID, notification_id: UUID
    ) -> NotificationReadResponse:
        """Mark a notification as read for a user.

        Args:
            user_id: The user's unique identifier
            notification_id: The notification's unique identifier

        Returns:
            NotificationReadResponse: Confirmation of marking as read

        Raises:
            HTTPException: If notification not found, not owned by user, or DB error
        """
        _log.info(f"Marking notification {notification_id} as read for user {user_id}")
        try:
            notification_result = await self.db.execute(
                select(NotificationModel).where(
                    NotificationModel.notification_id == notification_id,
                    NotificationModel.user_id == user_id,
                    NotificationModel.is_deleted.is_(False),
                )
            )
            notification = notification_result.scalar_one_or_none()
            if not notification:
                _log.warning(
                    f"Notification {notification_id} not found for user {user_id}"
                )
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="Notification not found",
                )
            if notification.is_read:
                _log.info(f"Notification {notification_id} already marked as read")
            else:
                notification.is_read = True
                await self.db.commit()
                await self.db.refresh(notification)
                _log.info(
                    f"Notification {notification_id} marked as read for user {user_id}"
                )
            return NotificationReadResponse(
                message="Notification marked as read successfully",
            )
        except DatabaseError as e:
            _log.error(f"Database error while marking notification read: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            _log.error(
                f"Database connection error while marking notification read: {e}"
            )
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Database service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e
