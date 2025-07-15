"""User logout response schemas."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserLogoutResponse(BaseSchemaModel):
    """Response schema for successful user logout."""

    message: str = Field(
        ...,
        description="Logout confirmation message",
        examples=["User logged out successfully"],
    )
    session_invalidated: bool = Field(
        ...,
        description="Whether the user session was successfully invalidated",
        examples=[True],
    )
