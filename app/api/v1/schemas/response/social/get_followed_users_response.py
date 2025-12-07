"""GetFollowedUsersResponse schema."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel
from app.api.v1.schemas.common.user import User


class GetFollowedUsersResponse(BaseSchemaModel):
    """Response schema for users the current user is following (no sensitive info)."""

    total_count: int = Field(
        ...,
        description="Total number of users being followed",
    )
    followed_users: list[User] | None = Field(
        None,
        description="List of users being followed (null when count_only=true)",
    )
    limit: int | None = Field(
        None,
        description="Number of results returned (null when count_only=true)",
    )
    offset: int | None = Field(
        None,
        description="Number of results skipped (null when count_only=true)",
    )
