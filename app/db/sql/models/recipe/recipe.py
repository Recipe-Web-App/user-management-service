"""SQLAlchemy model for the recipes table."""

from sqlalchemy import (
    TIMESTAMP,
    BigInteger,
    Column,
    Enum,
    ForeignKey,
    Integer,
    Numeric,
    String,
    Text,
    func,
)
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship

from app.core.config import settings
from app.db.sql.models.base_sql_model import BaseSqlModel
from app.enums.difficulty_level_enum import DifficultyLevelEnum


class Recipe(BaseSqlModel):
    """SQLAlchemy model for the recipes table."""

    __tablename__ = "recipes"
    __table_args__ = {"schema": settings.postgres_schema}  # noqa: RUF012

    recipe_id = Column(BigInteger, primary_key=True, autoincrement=True)
    user_id = Column(
        UUID(as_uuid=True),
        ForeignKey("recipe_manager.users.user_id", ondelete="CASCADE"),
        nullable=False,
    )
    title = Column(String(255), nullable=False)
    description = Column(Text)
    origin_url = Column(Text)
    servings = Column(Numeric(5, 2))
    preparation_time = Column(Integer)
    cooking_time = Column(Integer)
    difficulty: Column = Column(
        Enum(
            DifficultyLevelEnum,
            name="difficulty_level_enum",
            schema=settings.postgres_schema,
            create_constraint=False,
        ),
        nullable=True,
    )
    created_at = Column(TIMESTAMP(timezone=True), server_default=func.now())
    updated_at = Column(
        TIMESTAMP(timezone=True), server_default=func.now(), onupdate=func.now()
    )

    # Relationships (minimal, can be expanded as needed)
    user = relationship("User", back_populates="recipes")

    reviews = relationship(
        "RecipeReview",
        back_populates="recipe",
        cascade="all, delete-orphan",
        doc="List of reviews for this recipe.",
    )

    favorites = relationship(
        "RecipeFavorite",
        back_populates="recipe",
        cascade="all, delete-orphan",
        doc="List of users who have favorited this recipe.",
    )
