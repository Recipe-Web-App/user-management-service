"""User follows association table model."""

from datetime import UTC, datetime

from sqlalchemy import Column, DateTime, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship

from app.db.sql.models.base_sql_model import BaseSqlModel


class UserFollows(BaseSqlModel):
    """Association table for user follow relationships.

    Each row represents a directed follow relationship:
    - follower_id: the user who is following
    - followee_id: the user being followed
    - followed_at: when the follow occurred
    The composite primary key (follower_id, followee_id) ensures uniqueness.
    """

    __tablename__ = "user_follows"
    __table_args__ = {"schema": "recipe_manager"}  # noqa: RUF012

    follower_id = Column(
        UUID(as_uuid=True),
        ForeignKey("recipe_manager.users.user_id", ondelete="CASCADE"),
        primary_key=True,
    )
    followee_id = Column(
        UUID(as_uuid=True),
        ForeignKey("recipe_manager.users.user_id", ondelete="CASCADE"),
        primary_key=True,
    )
    followed_at = Column(
        DateTime(timezone=True), default=lambda: datetime.now(UTC), nullable=False
    )

    follower = relationship(
        "User", foreign_keys=[follower_id], back_populates="following_associations"
    )
    followee = relationship(
        "User", foreign_keys=[followee_id], back_populates="follower_associations"
    )
