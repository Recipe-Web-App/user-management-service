"""User display preferences model."""

from datetime import UTC, datetime
from uuid import UUID

from sqlalchemy import Boolean, DateTime, Enum, ForeignKey
from sqlalchemy.dialects.postgresql import UUID as PostgresUUID  # noqa: N811
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.db.sql.models.base_sql_model import BaseSqlModel
from app.enums.preferences.color_scheme_enum import ColorSchemeEnum
from app.enums.preferences.font_size_enum import FontSizeEnum
from app.enums.preferences.layout_density_enum import LayoutDensityEnum


class UserDisplayPreferences(BaseSqlModel):
    """SQLAlchemy model for user display preferences."""

    __tablename__ = "user_display_preferences"
    __table_args__ = {"schema": "recipe_manager"}  # noqa: RUF012

    id: Mapped[UUID] = mapped_column(PostgresUUID(as_uuid=True), primary_key=True)
    user_id: Mapped[UUID] = mapped_column(
        PostgresUUID(as_uuid=True),
        ForeignKey("recipe_manager.users.user_id", ondelete="CASCADE"),
        nullable=False,
        unique=True,
    )
    font_size: Mapped[str] = mapped_column(
        Enum(FontSizeEnum, name="font_size_enum", create_constraint=False),
        nullable=False,
        default=FontSizeEnum.MEDIUM,
    )
    color_scheme: Mapped[str] = mapped_column(
        Enum(ColorSchemeEnum, name="color_scheme_enum", create_constraint=False),
        nullable=False,
        default=ColorSchemeEnum.LIGHT,
    )
    layout_density: Mapped[str] = mapped_column(
        Enum(LayoutDensityEnum, name="layout_density_enum", create_constraint=False),
        nullable=False,
        default=LayoutDensityEnum.COMFORTABLE,
    )
    show_images: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    compact_mode: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, default=datetime.now(UTC)
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        nullable=False,
        default=datetime.now(UTC),
        onupdate=datetime.now(UTC),
    )

    user = relationship("User", back_populates="display_preferences")
