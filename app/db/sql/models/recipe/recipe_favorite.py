"""SQLAlchemy model for the recipe_favorites table."""

from sqlalchemy import TIMESTAMP, BigInteger, Column, ForeignKey, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship

from app.db.sql.models.base_sql_model import BaseSqlModel


class RecipeFavorite(BaseSqlModel):
    """SQLAlchemy model for the recipe_favorites table."""

    __tablename__ = "recipe_favorites"
    __table_args__ = {"schema": "recipe_manager"}  # noqa: RUF012

    user_id = Column(
        UUID(as_uuid=True),
        ForeignKey("recipe_manager.users.user_id", ondelete="CASCADE"),
        primary_key=True,
    )
    recipe_id = Column(
        BigInteger,
        ForeignKey("recipe_manager.recipes.recipe_id", ondelete="CASCADE"),
        primary_key=True,
    )
    favorited_at = Column(TIMESTAMP(timezone=True), server_default=func.now())

    user = relationship("User", back_populates="recipe_favorites")
    recipe = relationship("Recipe", back_populates="favorites")
