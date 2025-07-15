"""User password reset confirmation response schemas."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserPasswordResetConfirmResponse(BaseSchemaModel):
    """Response schema for password reset confirmation."""

    message: str = Field(
        ...,
        description="Success message",
        examples=["Password reset successfully"],
    )
    password_updated: bool = Field(
        ...,
        description="Whether the password was updated",
    )
