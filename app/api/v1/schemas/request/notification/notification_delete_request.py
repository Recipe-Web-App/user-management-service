"""Notification delete request schema."""

from typing import Annotated
from uuid import UUID

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class NotificationDeleteRequest(BaseSchemaModel):
    """Request schema for deleting notifications."""

    notification_ids: Annotated[list[UUID], Field(min_length=1, max_length=100)] = (
        Field(
            ...,
            description="List of notification IDs to delete",
        )
    )
