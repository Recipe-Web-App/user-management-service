"""Common user schema definitions."""

from datetime import datetime
from uuid import UUID

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class User(BaseSchemaModel):
    """Response schema for user data."""

    user_id: UUID = Field(
        ...,
        description="Unique identifier for the user",
    )
    username: str = Field(
        ...,
        description="Username for the user account",
        min_length=1,
        max_length=50,
    )
    email: str = Field(
        ...,
        description="Email address for the user account",
    )
    full_name: str | None = Field(
        None,
        description="Full name of the user",
        max_length=100,
    )
    bio: str | None = Field(
        None,
        description="User's biography or description",
        max_length=500,
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
