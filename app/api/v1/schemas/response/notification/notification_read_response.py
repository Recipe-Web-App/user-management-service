"""Notification read response schema."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class NotificationReadResponse(BaseSchemaModel):
    """Response schema for marking notification as read."""

    message: str = Field(
        ...,
        description="Success message",
        examples=["Notification marked as read successfully"],
    )
