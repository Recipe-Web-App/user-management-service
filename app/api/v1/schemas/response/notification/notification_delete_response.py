"""Notification delete response schema."""

from uuid import UUID

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class NotificationDeleteResponse(BaseSchemaModel):
    """Response schema for deleting notifications."""

    message: str = Field(
        ...,
        description="Success message",
        examples=["Notifications deleted successfully"],
    )
    deleted_notification_ids: list[UUID] = Field(
        ...,
        description="List of notification IDs that were successfully deleted",
    )
