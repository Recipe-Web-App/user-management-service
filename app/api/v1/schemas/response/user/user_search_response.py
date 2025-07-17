"""User search response schemas.

Defines response models for user search endpoints, including paginated results and user
summaries.
"""

from datetime import datetime
from uuid import UUID

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserSearchResult(BaseSchemaModel):
    """Summary of a user for search results."""

    user_id: UUID = Field(..., description="Unique identifier for the user")
    username: str = Field(..., description="Username for the user account")
    full_name: str | None = Field(None, description="Full name of the user")
    is_active: bool = Field(..., description="Whether the user account is active")
    created_at: datetime = Field(
        ..., description="Timestamp when the user account was created"
    )
    updated_at: datetime = Field(
        ..., description="Timestamp when the user account was last updated"
    )


class UserSearchResponse(BaseSchemaModel):
    """Paginated user search response."""

    results: list[UserSearchResult] = Field(
        ..., description="List of users matching the search query"
    )
    total_count: int = Field(
        ..., description="Total number of users matching the search query"
    )
    limit: int = Field(..., description="Number of results returned")
    offset: int = Field(..., description="Number of results skipped")
