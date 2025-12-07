"""Follow response schema."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class FollowResponse(BaseSchemaModel):
    """Response schema for follow/unfollow actions."""

    message: str = Field(
        ...,
        description="Success message for the follow/unfollow action",
    )
    is_following: bool = Field(
        ...,
        description="Whether the user is now following the target user",
    )
