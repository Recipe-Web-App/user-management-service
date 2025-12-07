"""User profile response schema definitions."""

from datetime import datetime
from uuid import UUID

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserProfileResponse(BaseSchemaModel):
    """Response schema for user profile data with privacy-controlled fields."""

    user_id: UUID = Field(
        ...,
        description="Unique identifier for the user",
    )
    username: str = Field(
        ...,
        description="Username for the user account",
    )
    email: str | None = Field(
        None,
        description=(
            "Email address (only included if contact info is visible to requester)"
        ),
    )
    full_name: str | None = Field(
        None,
        description="Full name of the user",
    )
    bio: str | None = Field(
        None,
        description="User's biography or description",
    )
    is_active: bool = Field(
        ...,
        description="Whether the user account is active",
    )
    created_at: datetime = Field(
        ...,
        description="Timestamp when the user account was created",
    )
    updated_at: datetime = Field(
        ...,
        description="Timestamp when the user account was last updated",
    )
