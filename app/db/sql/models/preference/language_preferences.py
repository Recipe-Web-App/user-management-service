"""User language preferences model."""

from datetime import UTC, datetime
from uuid import UUID

from sqlalchemy import Boolean, DateTime, ForeignKey
from sqlalchemy.dialects.postgresql import ENUM as SAEnum  # noqa: N811
from sqlalchemy.dialects.postgresql import UUID as PostgresUUID  # noqa: N811
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.db.sql.models.base_sql_model import BaseSqlModel
from app.enums.preferences.language_enum import LanguageEnum


class UserLanguagePreferences(BaseSqlModel):
    """SQLAlchemy model for user language preferences."""

    __tablename__ = "user_language_preferences"
    __table_args__ = {"schema": "recipe_manager"}  # noqa: RUF012

    id: Mapped[UUID] = mapped_column(PostgresUUID(as_uuid=True), primary_key=True)
    user_id: Mapped[UUID] = mapped_column(
        PostgresUUID(as_uuid=True),
        ForeignKey("recipe_manager.users.user_id", ondelete="CASCADE"),
        nullable=False,
        unique=True,
    )
    primary_language: Mapped[str] = mapped_column(
        SAEnum(
            LanguageEnum,
            name="language_enum",
            schema="recipe_manager",
            native_enum=False,
            create_constraint=False,
        ),
        nullable=False,
        default=LanguageEnum.EN,
    )
    secondary_language: Mapped[str] = mapped_column(
        SAEnum(
            LanguageEnum,
            name="language_enum",
            schema="recipe_manager",
            native_enum=False,
            create_constraint=False,
        ),
        nullable=True,
    )
    translation_enabled: Mapped[bool] = mapped_column(
        Boolean, nullable=False, default=False
    )
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, default=datetime.now(UTC)
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        nullable=False,
        default=datetime.now(UTC),
        onupdate=datetime.now(UTC),
    )

    user = relationship("User", back_populates="language_preferences")
