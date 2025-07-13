"""Session management service for user authentication."""

import uuid
from datetime import UTC, datetime, timedelta
from typing import Any

import redis.asyncio as redis
from pydantic import BaseModel, Field

from app.core.config import settings
from app.core.logging import get_logger

_log = get_logger(__name__)


class SessionData(BaseModel):
    """Model for session data structure."""

    user_id: str
    session_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    created_at: datetime = Field(default_factory=datetime.utcnow)
    expires_at: datetime
    is_active: bool = True
    metadata: dict[str, Any] = Field(default_factory=dict)
    last_activity: datetime = Field(default_factory=datetime.utcnow)


class SessionService:
    """Manages user sessions using Redis."""

    def __init__(self) -> None:
        """Initialize session service with Redis connection."""
        self.redis_client = redis.Redis(
            host=settings.redis_host,
            port=settings.redis_port,
            password=settings.redis_password,
            db=settings.redis_db,
            decode_responses=True,
        )
        self.session_prefix = "session:"
        self.user_sessions_prefix = "user_sessions:"
        self.session_cleanup_key = "session_cleanup"

    async def create_session(
        self,
        user_id: str,
        ttl_seconds: int = 3600,
        metadata: dict[str, Any] | None = None,
    ) -> SessionData:
        """Create a new session for a user.

        Args:
            user_id: The user ID for the session
            ttl_seconds: Session time-to-live in seconds
            metadata: Additional session metadata

        Returns:
            SessionData: The created session data
        """
        _log.info(f"Creating session for user: {user_id}")

        session_data = SessionData(
            user_id=user_id,
            expires_at=datetime.now(UTC) + timedelta(seconds=ttl_seconds),
            metadata=metadata or {},
        )

        # Store session data
        session_key = f"{self.session_prefix}{session_data.session_id}"
        await self.redis_client.setex(
            session_key, ttl_seconds, session_data.model_dump_json()
        )

        # Add to user's session list
        user_sessions_key = f"{self.user_sessions_prefix}{user_id}"
        await self.redis_client.sadd(user_sessions_key, session_data.session_id)
        await self.redis_client.expire(user_sessions_key, ttl_seconds)

        # Add to cleanup set
        await self.redis_client.zadd(
            self.session_cleanup_key,
            {session_data.session_id: session_data.expires_at.timestamp()},
        )

        _log.info(f"Session created: {session_data.session_id}")
        return session_data

    async def get_session(self, session_id: str) -> SessionData | None:
        """Retrieve session data by session ID.

        Args:
            session_id: The session ID to retrieve

        Returns:
            Optional[SessionData]: The session data if found and active
        """
        session_key = f"{self.session_prefix}{session_id}"
        session_json = await self.redis_client.get(session_key)

        if not session_json:
            _log.debug(f"Session not found: {session_id}")
            return None

        try:
            session_data = SessionData.model_validate_json(str(session_json))

            # Update last activity
            session_data.last_activity = datetime.now(UTC)
            remaining_ttl = await self._get_remaining_ttl(session_id)

            if remaining_ttl > 0:
                await self.redis_client.setex(
                    session_key,
                    remaining_ttl,
                    session_data.model_dump_json(),
                )
                _log.debug(f"Session retrieved: {session_id}")
                return session_data
            _log.debug(f"Session expired: {session_id}")
            await self.invalidate_session(session_id)
        except Exception as e:
            _log.error(f"Error parsing session data for {session_id}: {e}")
            return None
        else:
            return None

    async def invalidate_session(self, session_id: str) -> bool:
        """Invalidate a specific session.

        Args:
            session_id: The session ID to invalidate

        Returns:
            bool: True if session was invalidated successfully
        """
        session_key = f"{self.session_prefix}{session_id}"
        session_json = await self.redis_client.get(session_key)

        if not session_json:
            _log.debug(f"Session not found for invalidation: {session_id}")
            return False

        try:
            session_data = SessionData.model_validate_json(str(session_json))

            # Remove from Redis
            await self.redis_client.delete(session_key)
            await self.redis_client.zrem(self.session_cleanup_key, session_id)

            # Remove from user's session list
            user_sessions_key = f"{self.user_sessions_prefix}{session_data.user_id}"
            await self.redis_client.srem(user_sessions_key, session_id)

            _log.info(f"Session invalidated: {session_id}")
        except Exception as e:
            _log.error(f"Error invalidating session {session_id}: {e}")
            return False
        else:
            return True

    async def invalidate_user_sessions(self, user_id: str) -> int:
        """Invalidate all sessions for a user.

        Args:
            user_id: The user ID whose sessions to invalidate

        Returns:
            int: Number of sessions invalidated
        """
        user_sessions_key = f"{self.user_sessions_prefix}{user_id}"
        session_ids = await self.redis_client.smembers(user_sessions_key)

        if not session_ids:
            _log.debug(f"No sessions found for user: {user_id}")
            return 0

        # Remove all sessions
        for session_id in session_ids:
            session_key = f"{self.session_prefix}{session_id}"
            await self.redis_client.delete(session_key)
            await self.redis_client.zrem(self.session_cleanup_key, session_id)

        # Remove user sessions set
        await self.redis_client.delete(user_sessions_key)

        _log.info(f"Invalidated {len(session_ids)} sessions for user: {user_id}")
        return len(session_ids)

    async def get_user_sessions(self, user_id: str) -> list[SessionData]:
        """Get all active sessions for a user.

        Args:
            user_id: The user ID to get sessions for

        Returns:
            list[SessionData]: List of active sessions
        """
        user_sessions_key = f"{self.user_sessions_prefix}{user_id}"
        session_ids = await self.redis_client.smembers(user_sessions_key)

        sessions = []
        for session_id in session_ids:
            session_data = await self.get_session(session_id)
            if session_data and session_data.is_active:
                sessions.append(session_data)

        _log.debug(f"Found {len(sessions)} active sessions for user: {user_id}")
        return sessions

    async def cleanup_expired_sessions(self) -> int:
        """Clean up expired sessions.

        Returns:
            int: Number of sessions cleaned up
        """
        current_time = datetime.now(UTC).timestamp()
        expired_sessions = await self.redis_client.zrangebyscore(
            self.session_cleanup_key, 0, current_time
        )

        cleaned_count = 0
        for session_id in expired_sessions:
            session_id_str = session_id
            if await self.invalidate_session(session_id_str):
                cleaned_count += 1

        _log.info(f"Cleaned up {cleaned_count} expired sessions")
        return cleaned_count

    async def _get_remaining_ttl(self, session_id: str) -> int:
        """Get remaining TTL for a session.

        Args:
            session_id: The session ID

        Returns:
            int: Remaining TTL in seconds
        """
        session_key = f"{self.session_prefix}{session_id}"
        return await self.redis_client.ttl(session_key)

    async def get_session_stats(self) -> dict[str, Any]:
        """Get session statistics.

        Returns:
            Dict[str, Any]: Session statistics
        """
        total_sessions = await self.redis_client.zcard(self.session_cleanup_key)
        current_time = datetime.now(UTC).timestamp()
        active_sessions = await self.redis_client.zcount(
            self.session_cleanup_key, current_time, "+inf"
        )

        return {
            "total_sessions": total_sessions,
            "active_sessions": active_sessions,
            "expired_sessions": total_sessions - active_sessions,
        }

    async def ping(self) -> bool:
        """Test Redis connection.

        Returns:
            bool: True if connection is successful
        """
        try:
            return bool(await self.redis_client.ping())
        except redis.ConnectionError:
            return False
