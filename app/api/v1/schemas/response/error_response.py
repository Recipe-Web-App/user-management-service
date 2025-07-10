"""Error response schemas."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class ErrorResponse(BaseSchemaModel):
    """Error response schema."""

    error: str = Field(
        ...,
        description="Error type or category",
        examples=["validation_error", "authentication_error", "not_found"],
    )
    message: str = Field(
        ...,
        description="Human-readable error message",
        examples=["Invalid input data", "User not found", "Authentication failed"],
    )
    details: dict | None = Field(
        None,
        description="Additional error details or context",
        examples=[{"field": "email", "issue": "Invalid email format"}],
    )
