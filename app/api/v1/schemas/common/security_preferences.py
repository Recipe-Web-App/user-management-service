"""Security preferences schema definitions."""

from pydantic import BaseModel, Field


class SecurityPreferences(BaseModel):
    """Schema for user security preferences."""

    two_factor_auth: bool | None = Field(
        None, description="Whether two-factor authentication is enabled"
    )
    login_notifications: bool | None = Field(
        None, description="Whether login notifications are enabled"
    )
    session_timeout: bool | None = Field(
        None, description="Whether session timeout is enabled"
    )
    password_requirements: bool | None = Field(
        None, description="Whether password requirements are enforced"
    )
