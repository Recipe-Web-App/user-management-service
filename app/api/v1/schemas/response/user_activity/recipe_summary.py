"""RecipeSummary schema for user activity endpoint.

Represents a summary of a recipe created by the user, including id, title, and creation
time.
"""

from datetime import datetime

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class RecipeSummary(BaseSchemaModel):
    """Summary of a recipe for user activity feed."""

    recipe_id: int = Field(..., description="Unique identifier for the recipe.")
    title: str = Field(..., description="Title of the recipe.")
    created_at: datetime = Field(
        ..., description="Timestamp when the recipe was created."
    )
