"""Notification count response schema."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class NotificationCountResponse(BaseSchemaModel):
    """Response schema for notification count only."""

    total_count: int = Field(
        ...,
        description="Total number of notifications",
    )
