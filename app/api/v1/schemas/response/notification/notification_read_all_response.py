"""Notification read all response schema."""

from uuid import UUID

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class NotificationReadAllResponse(BaseSchemaModel):
    """Response schema for marking all notifications as read."""

    message: str = Field(
        ...,
        description="Success message",
        examples=["All notifications marked as read successfully"],
    )
    read_notification_ids: list[UUID] = Field(
        ...,
        description="List of notification IDs that were marked as read",
    )
