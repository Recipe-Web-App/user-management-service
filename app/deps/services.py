"""Service dependency providers."""

from typing import Annotated

from fastapi import Depends

from app.clients.notification_client import NotificationClient, get_notification_client
from app.deps.database import DatabaseSession, RedisDbSession
from app.services.admin_service import AdminService
from app.services.notification_service import NotificationService
from app.services.social_service import SocialService
from app.services.user_service import UserService


async def get_user_service(
    db_session: DatabaseSession,
    redis_session: RedisDbSession,
) -> UserService:
    """Get user service dependency."""
    return UserService(db_session, redis_session)


async def get_admin_service(
    db_session: DatabaseSession,
    redis_session: RedisDbSession,
) -> AdminService:
    """Get admin service dependency."""
    return AdminService(db_session, redis_session)


async def get_social_service(
    db_session: DatabaseSession,
    notification_client: Annotated[
        NotificationClient, Depends(get_notification_client)
    ],
) -> SocialService:
    """Get social service dependency.

    Args:
        db_session: Database session
        notification_client: Notification service client for sending notifications

    Returns:
        Configured SocialService instance
    """
    return SocialService(db_session, notification_client)


async def get_notification_service(
    db_session: DatabaseSession,
) -> NotificationService:
    """Get notification service dependency."""
    return NotificationService(db_session)


# Type aliases for dependency injection
UserServiceDep = Annotated[UserService, Depends(get_user_service)]
AdminServiceDep = Annotated[AdminService, Depends(get_admin_service)]
SocialServiceDep = Annotated[SocialService, Depends(get_social_service)]
NotificationServiceDep = Annotated[
    NotificationService, Depends(get_notification_service)
]
