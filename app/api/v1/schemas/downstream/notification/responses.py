"""Response schemas for notification service API."""

from pydantic import BaseModel, Field


class NotificationItem(BaseModel):
    """Individual notification item in batch response.

    Attributes:
        notification_id: UUID of the created notification
        recipient_id: UUID of the recipient user
    """

    notification_id: str = Field(
        ...,
        description="UUID of the created notification",
        examples=["770e8400-e29b-41d4-a716-446655440111"],
    )
    recipient_id: str = Field(
        ...,
        description="UUID of the recipient user",
        examples=["550e8400-e29b-41d4-a716-446655440001"],
    )


class BatchNotificationResponse(BaseModel):
    """Response from notification service batch operations.

    Returned by notification service when notifications are queued (202 Accepted).
    Matches the notification service API specification.

    Attributes:
        notifications: List of created notifications with their IDs
        queued_count: Number of notifications successfully queued
        message: Success message from the service
    """

    notifications: list[NotificationItem] = Field(
        ...,
        description="List of created notifications mapped to recipients",
    )
    queued_count: int = Field(
        ...,
        description="Number of notifications successfully queued",
        examples=[2],
    )
    message: str = Field(
        ...,
        description="Success message",
        examples=["Notifications queued successfully"],
    )
