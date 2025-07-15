"""User password reset request schemas."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserPasswordResetRequest(BaseSchemaModel):
    """Request schema for password reset."""

    email: str = Field(
        ...,
        max_length=255,
        description="Email address to send password reset to",
        examples=["user@example.com"],
    )
