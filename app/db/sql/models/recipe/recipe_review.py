"""SQLAlchemy model for the reviews table."""

from sqlalchemy import TIMESTAMP, BigInteger, Column, ForeignKey, Numeric, Text, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship

from app.db.sql.models.base_sql_model import BaseSqlModel


class RecipeReview(BaseSqlModel):
    """SQLAlchemy model for the reviews table."""

    __tablename__ = "reviews"
    __table_args__ = {"schema": "recipe_manager"}  # noqa: RUF012

    review_id = Column(BigInteger, primary_key=True, autoincrement=True)
    recipe_id = Column(
        BigInteger,
        ForeignKey("recipe_manager.recipes.recipe_id", ondelete="CASCADE"),
        nullable=False,
    )
    user_id = Column(
        UUID(as_uuid=True),
        ForeignKey("recipe_manager.users.user_id", ondelete="CASCADE"),
        nullable=False,
    )
    rating = Column(Numeric(2, 1), nullable=False)
    comment = Column(Text)
    created_at = Column(TIMESTAMP(timezone=True), server_default=func.now())

    recipe = relationship("Recipe", back_populates="reviews")
    user = relationship("User", back_populates="recipe_reviews")
