"""User activity response schema for the user activity endpoint.

Defines the UserActivityResponse model, which aggregates recent user activities such as
created recipes, followed users, reviews, and favorites, each limited by per_type_limit.
"""

from uuid import UUID

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel
from app.api.v1.schemas.response.user_activity.favorite_summary import FavoriteSummary
from app.api.v1.schemas.response.user_activity.recipe_summary import RecipeSummary
from app.api.v1.schemas.response.user_activity.review_summary import ReviewSummary
from app.api.v1.schemas.response.user_activity.user_summary import UserSummary


class UserActivityResponse(BaseSchemaModel):
    """Response schema for user activity endpoint, containing recent activity lists."""

    user_id: UUID = Field(..., description="Unique identifier for the user.")
    recent_recipes: list[RecipeSummary] = Field(
        ..., description="List of recently created recipes."
    )
    recent_follows: list[UserSummary] = Field(
        ..., description="List of recently followed users."
    )
    recent_reviews: list[ReviewSummary] = Field(
        ..., description="List of recent reviews written by the user."
    )
    recent_favorites: list[FavoriteSummary] = Field(
        ..., description="List of recently favorited recipes."
    )
    per_type_limit: int = Field(..., description="Limit applied to each activity type.")
