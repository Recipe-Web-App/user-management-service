"""FavoriteSummary schema for user activity endpoint.

Represents a summary of a recipe favorited by the user, including id, title, and
favorited time.
"""

from datetime import datetime

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class FavoriteSummary(BaseSchemaModel):
    """Summary of a favorited recipe for user activity feed."""

    recipe_id: int = Field(
        ..., description="Unique identifier for the favorited recipe."
    )
    title: str = Field(..., description="Title of the favorited recipe.")
    favorited_at: datetime = Field(
        ..., description="Timestamp when the recipe was favorited."
    )
