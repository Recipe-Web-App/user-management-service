"""Security preferences schema definitions."""

from pydantic import BaseModel, Field


class SecurityPreferences(BaseModel):
    """Schema for user security preferences."""

    two_factor_auth: bool = Field(
        ..., description="Whether two-factor authentication is enabled"
    )
    login_notifications: bool = Field(
        ..., description="Whether login notifications are enabled"
    )
    session_timeout: bool = Field(..., description="Whether session timeout is enabled")
    password_requirements: bool = Field(
        ..., description="Whether password requirements are enforced"
    )
