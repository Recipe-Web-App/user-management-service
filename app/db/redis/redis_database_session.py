"""Custom Redis session with domain-specific operations."""

from datetime import UTC, datetime, timedelta
from typing import Any

import redis.asyncio as redis

from app.core.logging import get_logger
from app.db.redis.models.session_data import SessionData

_log = get_logger(__name__)


class RedisDatabaseSession:
    """Custom Redis session with session management methods."""

    def __init__(self, redis_client: redis.Redis) -> None:
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
            redis.ConnectionError: If Redis connection fails
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
            sadd_result = self.redis.sadd(user_sessions_key, session_data.session_id)
            if hasattr(sadd_result, "__await__"):
                await sadd_result
            await self.redis.expire(user_sessions_key, ttl_seconds)

            # Add to cleanup set
            await self.redis.zadd(
                self.session_cleanup_key,
                {session_data.session_id: session_data.expires_at.timestamp()},
            )

            _log.info(f"Session created: {session_data.session_id}")
        except redis.ConnectionError as e:
            _log.error(
                f"Redis connection error while creating session for user {user_id}: {e}"
            )
            raise redis.ConnectionError(
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
            Optional[SessionData]: The session data if found and active

        Raises:
            redis.ConnectionError: If Redis connection fails
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
        except redis.ConnectionError as e:
            _log.error(
                f"Redis connection error while getting session {session_id}: {e}"
            )
            raise redis.ConnectionError(
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
            redis.ConnectionError: If Redis connection fails
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
            srem_result = self.redis.srem(user_sessions_key, session_id)
            if hasattr(srem_result, "__await__"):
                await srem_result

            _log.info(f"Session invalidated: {session_id}")
        except redis.ConnectionError as e:
            _log.error(
                f"Redis connection error while invalidating session {session_id}: {e}"
            )
            raise redis.ConnectionError(
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
            redis.ConnectionError: If Redis connection fails
        """
        try:
            user_sessions_key = f"{self.user_sessions_prefix}{user_id}"
            smembers_result = self.redis.smembers(user_sessions_key)
            if hasattr(smembers_result, "__await__"):
                session_ids = await smembers_result
            else:
                session_ids = smembers_result

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
        except redis.ConnectionError as e:
            _log.error(
                "Redis connection error while invalidating sessions for user "
                f"{user_id}: {e}"
            )
            raise redis.ConnectionError(
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
            redis.ConnectionError: If Redis connection fails
        """
        try:
            user_sessions_key = f"{self.user_sessions_prefix}{user_id}"
            smembers_result = self.redis.smembers(user_sessions_key)
            if hasattr(smembers_result, "__await__"):
                session_ids = await smembers_result
            else:
                session_ids = smembers_result

            sessions = []
            for session_id in session_ids:
                try:
                    session_data = await self.get_session(session_id)
                    if session_data and session_data.is_active:
                        sessions.append(session_data)
                except redis.ConnectionError:
                    # Skip this session if Redis is unavailable
                    continue

            _log.debug(f"Found {len(sessions)} active sessions for user: {user_id}")
        except redis.ConnectionError as e:
            _log.error(
                f"Redis connection error while getting sessions for user {user_id}: {e}"
            )
            raise redis.ConnectionError(
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
            redis.ConnectionError: If Redis connection fails
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
                except redis.ConnectionError:
                    # Skip this session if Redis is unavailable
                    continue

            _log.info(f"Cleaned up {cleaned_count} expired sessions")
        except redis.ConnectionError as e:
            _log.error(
                f"Redis connection error while cleaning up expired sessions: {e}"
            )
            raise redis.ConnectionError(
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
            redis.ConnectionError: If Redis connection fails
        """
        try:
            session_key = f"{self.session_prefix}{session_id}"
            return await self.redis.ttl(session_key)
        except redis.ConnectionError as e:
            _log.error(
                f"Redis connection error while getting TTL for session {session_id}: "
                f"{e}"
            )
            raise redis.ConnectionError(
                "Failed to get session TTL: Redis service unavailable"
            ) from e
        except Exception as e:
            _log.error(f"Error getting TTL for session {session_id}: {e}")
            return -1

    async def get_session_stats(self) -> dict[str, Any]:
        """Get session statistics.

        Returns:
            Dict[str, Any]: Session statistics

        Raises:
            redis.ConnectionError: If Redis connection fails
        """
        try:
            total_sessions = await self.redis.zcard(self.session_cleanup_key)
            current_time = datetime.now(UTC).timestamp()
            active_sessions = await self.redis.zcount(
                self.session_cleanup_key, current_time, "+inf"
            )

            return {
                "total_sessions": total_sessions,
                "active_sessions": active_sessions,
                "expired_sessions": total_sessions - active_sessions,
            }
        except redis.ConnectionError as e:
            _log.error(f"Redis connection error while getting session stats: {e}")
            raise redis.ConnectionError(
                "Failed to get session statistics: Redis service unavailable"
            ) from e
        except Exception as e:
            _log.error(f"Error getting session stats: {e}")
            return {
                "total_sessions": 0,
                "active_sessions": 0,
                "expired_sessions": 0,
            }

    async def ping(self) -> bool:
        """Test Redis connection.

        Returns:
            bool: True if connection is successful
        """
        try:
            return bool(await self.redis.ping())
        except redis.ConnectionError:
            return False

    # --- Deletion Token Management ---
    deletion_token_prefix = "deletion_token:"  # nosec B105
    deletion_token_ttl_seconds = 24 * 3600  # 24 hours

    async def store_deletion_token(
        self, user_id: str, token: str, expires_at: datetime
    ) -> None:
        """Store a deletion confirmation token for a user with expiration."""
        key = f"{self.deletion_token_prefix}{user_id}"
        await self.redis.hset(
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
        data = await self.redis.hgetall(key)
        return data if data else None

    async def delete_deletion_token(self, user_id: str) -> None:
        """Delete the deletion confirmation token for a user."""
        key = f"{self.deletion_token_prefix}{user_id}"
        await self.redis.delete(key)
