"""Notification list response schema."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel
from app.api.v1.schemas.common.notification import Notification


class NotificationListResponse(BaseSchemaModel):
    """Response schema for notification list."""

    notifications: list[Notification] = Field(
        ...,
        description="List of notifications",
    )
    total_count: int = Field(
        ...,
        description="Total number of notifications",
    )
    limit: int = Field(
        ...,
        description="Number of results returned",
    )
    offset: int = Field(
        ...,
        description="Number of results skipped",
    )
