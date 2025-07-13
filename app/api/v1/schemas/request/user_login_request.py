"""User login request schemas."""

from pydantic import Field, SecretStr, field_validator, model_validator

from app.api.v1.schemas.base_schema_model import BaseSchemaModel
from app.utils.constants import CONSTANTS


class UserLoginRequest(BaseSchemaModel):
    """Request schema for user login."""

    username: str | None = Field(
        None,
        min_length=1,
        max_length=50,
        description="Username for login (either username or email must be provided)",
        examples=["john_doe"],
    )
    email: str | None = Field(
        None,
        max_length=255,
        description=(
            "Email address for login (either username or email must be provided)"
        ),
        examples=["john@example.com"],
    )
    password: SecretStr = Field(
        ...,
        min_length=CONSTANTS.MIN_PASSWORD_LENGTH,
        max_length=CONSTANTS.MAX_PASSWORD_LENGTH,
        description="User password (will be redacted in logs)",
        examples=["SecurePass123!"],
    )

    @field_validator("username")
    @classmethod
    def validate_username(cls, v: str | None) -> str | None:
        """Validate username format if provided."""
        _ = cls  # Avoids vulture error

        if v is not None:
            if not v.strip():
                raise ValueError("Username cannot be empty")
            return v.strip().lower()
        return v

    @field_validator("email")
    @classmethod
    def validate_email(cls, v: str | None) -> str | None:
        """Validate email format if provided."""
        _ = cls  # Avoids vulture error

        if v is not None:
            if not v.strip():
                raise ValueError("Email cannot be empty")
            return v.strip().lower()
        return v

    @field_validator("password")
    @classmethod
    def validate_password(cls, v: SecretStr) -> SecretStr:
        """Validate password is not empty."""
        _ = cls  # Avoids vulture error

        password = v.get_secret_value()
        if not password.strip():
            raise ValueError("Password cannot be empty")
        return v

    @model_validator(mode="after")
    def validate_login_credentials(self) -> "UserLoginRequest":
        """Validate that either username or email is provided, but not both."""
        if not self.username and not self.email:
            raise ValueError("Either username or email must be provided")
        if self.username and self.email:
            raise ValueError("Provide either username or email, not both")
        return self
