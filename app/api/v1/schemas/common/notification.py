"""Common notification schema definitions."""

from datetime import datetime
from uuid import UUID

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class Notification(BaseSchemaModel):
    """Response schema for notification data."""

    notification_id: UUID = Field(
        ...,
        description="Unique identifier for the notification",
    )
    user_id: UUID = Field(
        ...,
        description="User ID who owns this notification",
    )
    title: str = Field(
        ...,
        description="Notification title",
        max_length=255,
    )
    message: str = Field(
        ...,
        description="Notification message content",
    )
    notification_type: str = Field(
        ...,
        description="Type of notification",
        max_length=50,
    )
    is_read: bool = Field(
        ...,
        description="Whether the notification has been read",
    )
    is_deleted: bool = Field(
        ...,
        description="Whether the notification has been deleted",
    )
    created_at: datetime = Field(
        ...,
        description="Timestamp when the notification was created",
    )
    updated_at: datetime = Field(
        ...,
        description="Timestamp when the notification was last updated",
    )
