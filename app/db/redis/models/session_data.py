"""Session data model for Redis storage."""

import uuid
from datetime import UTC, datetime, timedelta

from pydantic import Field

from app.db.redis.models.base_redis_model import BaseRedisModel


class SessionData(BaseRedisModel):
    """Model for session data structure."""

    user_id: str
    session_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    expires_at: datetime
    is_active: bool = True
    last_activity: datetime = Field(default_factory=lambda: datetime.now(UTC))

    def is_expired(self) -> bool:
        """Check if session is expired."""
        return datetime.now(UTC) > self.expires_at

    def extend_session(self, ttl_seconds: int) -> None:
        """Extend session expiration."""
        self.expires_at = datetime.now(UTC) + timedelta(seconds=ttl_seconds)
        self.update_timestamp()

    def update_activity(self) -> None:
        """Update last activity timestamp."""
        self.last_activity = datetime.now(UTC)
        self.update_timestamp()
