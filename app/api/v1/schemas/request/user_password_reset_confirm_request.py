"""User password reset confirmation request schemas."""

from pydantic import Field, SecretStr

from app.api.v1.schemas.base_schema_model import BaseSchemaModel
from app.utils.constants import CONSTANTS


class UserPasswordResetConfirmRequest(BaseSchemaModel):
    """Request schema for password reset confirmation."""

    reset_token: str = Field(
        ...,
        description="Password reset token received via email",
        examples=["eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."],
    )
    new_password: SecretStr = Field(
        ...,
        min_length=CONSTANTS.MIN_PASSWORD_LENGTH,
        max_length=CONSTANTS.MAX_PASSWORD_LENGTH,
        description="New password (will be redacted in logs)",
        examples=["NewSecurePass123!"],
    )
