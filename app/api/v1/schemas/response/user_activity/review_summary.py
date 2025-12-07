"""ReviewSummary schema for user activity endpoint.

Represents a summary of a review written by the user, including id, recipe, rating,
comment, and creation time.
"""

from datetime import datetime

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class ReviewSummary(BaseSchemaModel):
    """Summary of a recipe review for user activity feed."""

    review_id: int = Field(..., description="Unique identifier for the review.")
    recipe_id: int = Field(
        ..., description="Unique identifier for the reviewed recipe."
    )
    rating: float = Field(..., description="Rating given in the review.")
    comment: str | None = Field(None, description="Optional comment for the review.")
    created_at: datetime = Field(
        ..., description="Timestamp when the review was created."
    )
