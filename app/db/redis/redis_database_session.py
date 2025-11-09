"""Custom Redis session with domain-specific operations."""

from datetime import UTC, datetime, timedelta
from typing import Any

import redis.exceptions
from redis.asyncio import Redis as AsyncRedis

from app.api.v1.schemas.response.admin.redis_session_stats_response import (
    RedisSessionStatsResponse,
)
from app.core.logging import get_logger
from app.db.redis.models.session_data import SessionData

_log = get_logger(__name__)


class RedisDatabaseSession:
    """Custom Redis session with session management methods."""

    def __init__(self, redis_client: AsyncRedis) -> None:
        """Initialize Redis database session with Redis client.

        Args:
            redis_client: The Redis client instance to use for operations
        """
        self.redis = redis_client
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

        Raises:
            redis.exceptions.ConnectionError: If Redis connection fails
        """
        _log.info(f"Creating session for user: {user_id}")

        session_data = SessionData(
            user_id=user_id,
            expires_at=datetime.now(UTC) + timedelta(seconds=ttl_seconds),
            metadata=metadata or {},
        )

        try:
            # Store session data
            session_key = f"{self.session_prefix}{session_data.session_id}"
            await self.redis.setex(
                session_key, ttl_seconds, session_data.model_dump_json()
            )

            # Add to user's session list
            user_sessions_key = f"{self.user_sessions_prefix}{user_id}"
            await self.redis.sadd(user_sessions_key, session_data.session_id)  # type: ignore[misc]
            await self.redis.expire(user_sessions_key, ttl_seconds)

            # Add to cleanup set
            await self.redis.zadd(
                self.session_cleanup_key,
                {session_data.session_id: session_data.expires_at.timestamp()},
            )

            _log.info(f"Session created: {session_data.session_id}")
        except redis.exceptions.ConnectionError as e:
            _log.error(
                f"Redis connection error while creating session for user {user_id}: {e}"
            )
            raise redis.exceptions.ConnectionError(
                "Failed to create session: Redis service unavailable"
            ) from e
        except Exception as e:
            _log.error(f"Unexpected error creating session for user {user_id}: {e}")
            raise
        else:
            return session_data

    async def get_session(self, session_id: str) -> SessionData | None:
        """Retrieve session data by session ID.

        Args:
            session_id: The session ID to retrieve

        Returns:
            SessionData | None: The session data if found and active

        Raises:
            redis.exceptions.ConnectionError: If Redis connection fails
        """
        try:
            session_key = f"{self.session_prefix}{session_id}"
            session_json = await self.redis.get(session_key)

            if not session_json:
                _log.debug(f"Session not found: {session_id}")
                return None

            session_data = SessionData.model_validate_json(str(session_json))

            # Update last activity
            session_data.update_activity()
            remaining_ttl = await self._get_remaining_ttl(session_id)

            if remaining_ttl > 0:
                await self.redis.setex(
                    session_key,
                    remaining_ttl,
                    session_data.model_dump_json(),
                )
                _log.debug(f"Session retrieved: {session_id}")
                return session_data
            _log.debug(f"Session expired: {session_id}")
            await self.invalidate_session(session_id)
        except redis.exceptions.ConnectionError as e:
            _log.error(
                f"Redis connection error while getting session {session_id}: {e}"
            )
            raise redis.exceptions.ConnectionError(
                "Failed to retrieve session: Redis service unavailable"
            ) from e
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

        Raises:
            redis.exceptions.ConnectionError: If Redis connection fails
        """
        try:
            session_key = f"{self.session_prefix}{session_id}"
            session_json = await self.redis.get(session_key)

            if not session_json:
                _log.debug(f"Session not found for invalidation: {session_id}")
                return False

            session_data = SessionData.model_validate_json(str(session_json))

            # Remove from Redis
            await self.redis.delete(session_key)
            await self.redis.zrem(self.session_cleanup_key, session_id)

            # Remove from user's session list
            user_sessions_key = f"{self.user_sessions_prefix}{session_data.user_id}"
            await self.redis.srem(user_sessions_key, session_id)  # type: ignore[misc]

            _log.info(f"Session invalidated: {session_id}")
        except redis.exceptions.ConnectionError as e:
            _log.error(
                f"Redis connection error while invalidating session {session_id}: {e}"
            )
            raise redis.exceptions.ConnectionError(
                "Failed to invalidate session: Redis service unavailable"
            ) from e
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

        Raises:
            redis.exceptions.ConnectionError: If Redis connection fails
        """
        try:
            user_sessions_key = f"{self.user_sessions_prefix}{user_id}"
            session_ids = await self.redis.smembers(user_sessions_key)  # type: ignore[misc]

            if not session_ids:
                _log.debug(f"No sessions found for user: {user_id}")
                return 0

            # Remove all sessions
            for session_id in session_ids:
                session_key = f"{self.session_prefix}{session_id}"
                await self.redis.delete(session_key)
                await self.redis.zrem(self.session_cleanup_key, session_id)

            # Remove user sessions set
            await self.redis.delete(user_sessions_key)

            _log.info(f"Invalidated {len(session_ids)} sessions for user: {user_id}")
            return len(session_ids)
        except redis.exceptions.ConnectionError as e:
            _log.error(
                "Redis connection error while invalidating sessions for user "
                f"{user_id}: {e}"
            )
            raise redis.exceptions.ConnectionError(
                "Failed to invalidate user sessions: Redis service unavailable"
            ) from e
        except Exception as e:
            _log.error(f"Error invalidating sessions for user {user_id}: {e}")
            return 0

    async def get_user_sessions(self, user_id: str) -> list[SessionData]:
        """Get all active sessions for a user.

        Args:
            user_id: The user ID to get sessions for

        Returns:
            list[SessionData]: List of active sessions

        Raises:
            redis.exceptions.ConnectionError: If Redis connection fails
        """
        try:
            user_sessions_key = f"{self.user_sessions_prefix}{user_id}"
            session_ids = await self.redis.smembers(user_sessions_key)  # type: ignore[misc]

            sessions = []
            for session_id in session_ids:
                try:
                    session_data = await self.get_session(session_id)
                    if session_data and session_data.is_active:
                        sessions.append(session_data)
                except redis.exceptions.ConnectionError:
                    # Skip this session if Redis is unavailable
                    continue

            _log.debug(f"Found {len(sessions)} active sessions for user: {user_id}")
        except redis.exceptions.ConnectionError as e:
            _log.error(
                f"Redis connection error while getting sessions for user {user_id}: {e}"
            )
            raise redis.exceptions.ConnectionError(
                "Failed to get user sessions: Redis service unavailable"
            ) from e
        except Exception as e:
            _log.error(f"Error getting sessions for user {user_id}: {e}")
            return []
        else:
            return sessions

    async def cleanup_expired_sessions(self) -> int:
        """Clean up expired sessions.

        Returns:
            int: Number of sessions cleaned up

        Raises:
            redis.exceptions.ConnectionError: If Redis connection fails
        """
        try:
            current_time = datetime.now(UTC).timestamp()
            expired_sessions = await self.redis.zrangebyscore(
                self.session_cleanup_key, 0, current_time
            )

            cleaned_count = 0
            for session_id in expired_sessions:
                session_id_str = session_id
                try:
                    if await self.invalidate_session(session_id_str):
                        cleaned_count += 1
                except redis.exceptions.ConnectionError:
                    # Skip this session if Redis is unavailable
                    continue

            _log.info(f"Cleaned up {cleaned_count} expired sessions")
        except redis.exceptions.ConnectionError as e:
            _log.error(
                f"Redis connection error while cleaning up expired sessions: {e}"
            )
            raise redis.exceptions.ConnectionError(
                "Failed to cleanup expired sessions: Redis service unavailable"
            ) from e
        except Exception as e:
            _log.error(f"Error cleaning up expired sessions: {e}")
            return 0
        else:
            return cleaned_count

    async def _get_remaining_ttl(self, session_id: str) -> int:
        """Get remaining TTL for a session.

        Args:
            session_id: The session ID

        Returns:
            int: Remaining TTL in seconds

        Raises:
            redis.exceptions.ConnectionError: If Redis connection fails
        """
        try:
            session_key = f"{self.session_prefix}{session_id}"
            return await self.redis.ttl(session_key)
        except redis.exceptions.ConnectionError as e:
            _log.error(
                f"Redis connection error while getting TTL for session {session_id}: "
                f"{e}"
            )
            raise redis.exceptions.ConnectionError(
                "Failed to get session TTL: Redis service unavailable"
            ) from e
        except Exception as e:
            _log.error(f"Error getting TTL for session {session_id}: {e}")
            return -1

    async def get_session_stats(self) -> RedisSessionStatsResponse:
        """Get comprehensive session and Redis statistics.

        Returns:
            RedisSessionStatsResponse: All session and Redis stats
        Raises:
            redis.exceptions.ConnectionError: If Redis connection fails
        """
        try:
            # Session stats from sorted set
            total_sessions = await self.redis.zcard(self.session_cleanup_key)
            current_time = datetime.now(UTC).timestamp()
            active_sessions = await self.redis.zcount(
                self.session_cleanup_key, current_time, "+inf"
            )
            expired_sessions = total_sessions - active_sessions

            # Redis memory and key stats
            info = await self.redis.info()
            memory_usage_bytes = info.get("used_memory", 0)
            memory_usage_mb = round(memory_usage_bytes / (1024 * 1024), 2)
            key_count = await self.redis.dbsize()

            # Session key count
            session_keys = await self.redis.keys(f"{self.session_prefix}*")
            session_key_count = len(session_keys)

            # TTL stats for session keys
            ttls = []
            for key in session_keys:
                ttl = await self.redis.ttl(key)
                if ttl and ttl > 0:
                    ttls.append(ttl)
            session_ttl_min = min(ttls) if ttls else None
            session_ttl_max = max(ttls) if ttls else None
            session_ttl_avg = round(sum(ttls) / len(ttls), 2) if ttls else None

            # Redis server info
            redis_uptime_seconds = info.get("uptime_in_seconds", 0)
            redis_version = info.get("redis_version", "unknown")

            return RedisSessionStatsResponse(
                total_sessions=total_sessions,
                active_sessions=active_sessions,
                expired_sessions=expired_sessions,
                memory_usage_bytes=memory_usage_bytes,
                memory_usage_mb=memory_usage_mb,
                key_count=key_count,
                session_key_count=session_key_count,
                session_ttl_min=session_ttl_min,
                session_ttl_max=session_ttl_max,
                session_ttl_avg=session_ttl_avg,
                redis_uptime_seconds=redis_uptime_seconds,
                redis_version=redis_version,
            )
        except redis.exceptions.ConnectionError as e:
            _log.error(f"Redis connection error while getting session stats: {e}")
            raise redis.exceptions.ConnectionError(
                "Failed to get session statistics: Redis service unavailable"
            ) from e
        except Exception as e:
            _log.error(f"Error getting session stats: {e}")
            return RedisSessionStatsResponse(
                total_sessions=0,
                active_sessions=0,
                expired_sessions=0,
                memory_usage_bytes=0,
                memory_usage_mb=0.0,
                key_count=0,
                session_key_count=0,
                session_ttl_min=None,
                session_ttl_max=None,
                session_ttl_avg=None,
                redis_uptime_seconds=0,
                redis_version="unknown",
            )

    async def ping(self) -> bool:
        """Test Redis connection.

        Returns:
            bool: True if connection is successful
        """
        try:
            return bool(await self.redis.ping())  # type: ignore[misc]
        except redis.exceptions.ConnectionError:
            return False

    # --- Deletion Token Management ---
    deletion_token_prefix = "deletion_token:"  # nosec B105
    deletion_token_ttl_seconds = 24 * 3600  # 24 hours

    async def store_deletion_token(
        self, user_id: str, token: str, expires_at: datetime
    ) -> None:
        """Store a deletion confirmation token for a user with expiration."""
        key = f"{self.deletion_token_prefix}{user_id}"
        await self.redis.hset(  # type: ignore[misc]
            key,
            mapping={
                "token": token,
                "expires_at": expires_at.isoformat(),
                "user_id": user_id,
            },
        )
        await self.redis.expire(key, self.deletion_token_ttl_seconds)

    async def get_deletion_token(self, user_id: str) -> dict[str, str] | None:
        """Retrieve the deletion confirmation token for a user."""
        key = f"{self.deletion_token_prefix}{user_id}"
        data = await self.redis.hgetall(key)  # type: ignore[misc]
        return data or None

    async def delete_deletion_token(self, user_id: str) -> None:
        """Delete the deletion confirmation token for a user."""
        key = f"{self.deletion_token_prefix}{user_id}"
        await self.redis.delete(key)

    async def clear_all_sessions(self) -> int:
        """Clear all Redis sessions.

        Returns:
            int: Number of sessions cleared

        Raises:
            redis.exceptions.ConnectionError: If Redis connection fails
        """
        try:
            # Get all session keys
            session_keys = await self.redis.keys(f"{self.session_prefix}*")
            user_session_keys = await self.redis.keys(f"{self.user_sessions_prefix}*")

            total_keys = len(session_keys) + len(user_session_keys)

            if total_keys == 0:
                _log.info("No sessions to clear")
                return 0

            # Clear all session data
            if session_keys:
                await self.redis.delete(*session_keys)

            # Clear all user session lists
            if user_session_keys:
                await self.redis.delete(*user_session_keys)

            # Clear the cleanup sorted set
            await self.redis.delete(self.session_cleanup_key)

            _log.info(f"Cleared {total_keys} session-related keys from Redis")
        except redis.exceptions.ConnectionError as e:
            _log.error(f"Redis connection error while clearing all sessions: {e}")
            raise redis.exceptions.ConnectionError(
                "Failed to clear sessions: Redis service unavailable"
            ) from e
        except Exception as e:
            _log.error(f"Error clearing all sessions: {e}")
            raise
        else:
            return total_keys
