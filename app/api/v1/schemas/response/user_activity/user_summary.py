"""UserSummary schema for user activity endpoint.

Represents a summary of a user that was recently followed, including id, username, and
follow time.
"""

from datetime import datetime
from uuid import UUID

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserSummary(BaseSchemaModel):
    """Summary of a followed user for user activity feed."""

    user_id: UUID = Field(..., description="Unique identifier for the user.")
    username: str = Field(..., description="Username of the followed user.")
    followed_at: datetime = Field(
        ..., description="Timestamp when the user was followed."
    )
