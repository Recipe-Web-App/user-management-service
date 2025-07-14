"""User password reset response schemas."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserPasswordResetResponse(BaseSchemaModel):
    """Response schema for password reset request."""

    message: str = Field(
        ...,
        description="Success message",
        examples=["Password reset email sent successfully"],
    )
    email_sent: bool = Field(
        ...,
        description="Whether the reset email was sent",
    )
