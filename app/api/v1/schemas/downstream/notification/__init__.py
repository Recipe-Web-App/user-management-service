"""Schemas for notification service downstream API."""

from app.api.v1.schemas.downstream.notification.requests import (
    NewFollowerNotificationRequest,
)
from app.api.v1.schemas.downstream.notification.responses import (
    BatchNotificationResponse,
    NotificationItem,
)

__all__ = [
    "BatchNotificationResponse",
    "NewFollowerNotificationRequest",
    "NotificationItem",
]
