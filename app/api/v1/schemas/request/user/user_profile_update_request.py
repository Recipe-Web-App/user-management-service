"""User profile update request schema definitions."""

from pydantic import EmailStr, Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserProfileUpdateRequest(BaseSchemaModel):
    """Request schema for updating user profile information."""

    username: str | None = Field(
        None,
        min_length=3,
        max_length=50,
        pattern=r"^[a-zA-Z0-9_]+$",
        description=(
            "Unique username (3-50 characters, alphanumeric and underscore only)"
        ),
    )
    email: EmailStr | None = Field(None, description="Valid email address")
    full_name: str | None = Field(
        None,
        max_length=255,
        description="User's full name (max 255 characters)",
    )
    bio: str | None = Field(
        None,
        max_length=1000,
        description="User's bio/description (max 1000 characters)",
    )
