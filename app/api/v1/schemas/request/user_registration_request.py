"""User registration request schemas."""

from pydantic import EmailStr, Field, SecretStr, field_validator

from app.api.v1.schemas.base_schema_model import BaseSchemaModel
from app.utils.constants import CONSTANTS


class UserRegistrationRequest(BaseSchemaModel):
    """Request schema for user registration."""

    username: str = Field(
        ...,
        min_length=3,
        max_length=50,
        description="Username for the user account",
        examples=["john_doe", "jane_smith"],
    )
    email: EmailStr = Field(
        ...,
        description="User email address",
        examples=["john@example.com", "jane@example.com"],
    )
    password: SecretStr = Field(
        ...,
        min_length=CONSTANTS.MIN_PASSWORD_LENGTH,
        max_length=CONSTANTS.MAX_PASSWORD_LENGTH,
        description="User password (will be redacted in logs)",
        examples=["SecurePass123!"],
    )
    full_name: str | None = Field(
        None,
        max_length=255,
        description="Full name of the user",
        examples=["John Doe", "Jane Smith"],
    )
    bio: str | None = Field(
        None,
        description="User bio or description",
        examples=["Software developer", "Passionate about technology"],
    )

    @field_validator("username")
    @classmethod
    def validate_username(cls, v: str) -> str:
        """Validate username format."""
        _ = cls  # Avoids vulture error

        if not v.replace("_", "").replace("-", "").isalnum():
            raise ValueError(
                "Username must contain only letters, numbers, underscores, and hyphens"
            )
        return v.lower()

    @field_validator("password")
    @classmethod
    def validate_password(cls, v: SecretStr) -> SecretStr:
        """Validate password strength."""
        _ = cls  # Avoids vulture error

        password = v.get_secret_value()
        if len(password) < CONSTANTS.MIN_PASSWORD_LENGTH:
            raise ValueError(
                f"Password must be at least "
                f"{CONSTANTS.MIN_PASSWORD_LENGTH} characters long"
            )
        if len(password) > CONSTANTS.MAX_PASSWORD_LENGTH:
            raise ValueError(
                f"Password must be less than "
                f"{CONSTANTS.MAX_PASSWORD_LENGTH} characters long"
            )
        if not any(c.isupper() for c in password):
            raise ValueError("Password must contain at least one uppercase letter")
        if not any(c.islower() for c in password):
            raise ValueError("Password must contain at least one lowercase letter")
        if not any(c.isdigit() for c in password):
            raise ValueError("Password must contain at least one digit")
        return v
