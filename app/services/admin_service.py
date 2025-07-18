"""Admin service for operational and privileged admin endpoints."""

from datetime import UTC, datetime
from uuid import UUID

import redis
from fastapi import HTTPException, status

from app.api.v1.schemas.response.admin.clear_sessions_response import (
    ClearSessionsResponse,
)
from app.api.v1.schemas.response.admin.force_logout_response import ForceLogoutResponse
from app.api.v1.schemas.response.admin.redis_session_stats_response import (
    RedisSessionStatsResponse,
)
from app.api.v1.schemas.response.admin.system_health_response import (
    SystemHealthResponse,
)
from app.api.v1.schemas.response.admin.user_stats_response import UserStatsResponse
from app.core.config import settings
from app.core.logging import get_logger
from app.db.redis.redis_database_session import RedisDatabaseSession
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.enums.environment_enum import EnvironmentEnum
from app.enums.user_role_enum import UserRoleEnum
from app.exceptions.custom_exceptions.database_exceptions import DatabaseError

_log = get_logger(__name__)


class AdminService:
    """Service for admin operations."""

    def __init__(self, db: SqlDatabaseSession, redis: RedisDatabaseSession) -> None:
        """Initialize admin service with database and redis sessions."""
        self.db = db
        self.redis = redis

    async def _ensure_admin(self, user_id: str | UUID) -> None:
        """Ensure the user is an admin, else raise 403.

        Args:
            user_id: The user ID to check for admin privileges

        Raises:
            HTTPException: If user is not an admin
        """
        user_id_str = str(user_id)
        _log.debug(f"Checking admin privileges for user: {user_id_str}")

        try:
            user = await self.db.get_user_by_id(user_id_str)
            if not user:
                _log.warning(f"Admin check failed: User {user_id_str} not found")
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Admin privileges required.",
                )

            if user.role != UserRoleEnum.ADMIN:
                _log.warning(
                    f"Admin check failed: User {user_id_str} has role {user.role}"
                )
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Admin privileges required.",
                )

            _log.info(f"Admin privileges confirmed for user: {user_id_str}")
        except DatabaseError as e:
            _log.error(f"Database error: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e

    async def get_redis_session_stats(
        self, admin_user_id: str | UUID
    ) -> RedisSessionStatsResponse:
        """Return Redis session statistics using RedisDatabaseSession method.

        Args:
            admin_user_id: The admin user's ID (for admin check)
        Returns:
            RedisSessionStatsResponse: Redis session stats
        """
        _log.info(f"Admin {admin_user_id} requested Redis session stats")

        try:
            await self._ensure_admin(admin_user_id)
            stats = await self.redis.get_session_stats()
            _log.info(
                f"Redis session stats retrieved successfully for admin {admin_user_id}"
            )
        except redis.ConnectionError as e:
            _log.error(f"Redis connection error: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="Redis service unavailable.",
            ) from e
        else:
            return stats

    async def get_user_stats(self, admin_user_id: str | UUID) -> UserStatsResponse:
        """Return user statistics.

        Args:
            admin_user_id: The admin user's ID (for admin check)
        Returns:
            UserStatsResponse: User statistics
        """
        _log.info(f"Admin {admin_user_id} requested user statistics")

        try:
            await self._ensure_admin(admin_user_id)

            # Get user statistics from database
            user_stats = await self.db.get_user_statistics()
            users_online = await self.db.get_users_online_count()
            retention_rate = await self.db.get_user_retention_rate()
            avg_registration_rate = await self.db.get_average_registration_rate()

            response = UserStatsResponse(
                total_users=user_stats["total_users"],
                active_users=user_stats["active_users"],
                recently_registered=user_stats["recently_registered"],
                deactivated_users=user_stats["deactivated_users"],
                verified_users=user_stats["verified_users"],
                admin_users=user_stats["admin_users"],
                users_online=users_online,
                average_registration_rate=avg_registration_rate,
                user_retention_rate=retention_rate,
            )

            _log.info(
                f"User statistics retrieved successfully for admin {admin_user_id}: "
                f"total={response.total_users}, active={response.active_users}"
            )
        except DatabaseError as e:
            _log.error(f"Database error: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        else:
            return response

    async def get_system_health(
        self, admin_user_id: str | UUID
    ) -> SystemHealthResponse:
        """Return system health status.

        Args:
            admin_user_id: The admin user's ID (for admin check)
        Returns:
            SystemHealthResponse: System health status
        """
        _log.info(f"Admin {admin_user_id} requested system health check")

        await self._ensure_admin(admin_user_id)

        # Check database connection
        try:
            await self.db.execute("SELECT 1")
            database_status = True
            _log.debug("Database health check: OK")
        except Exception as e:
            database_status = False
            _log.error(f"Database health check failed: {e}")

        # Check Redis connection
        try:
            redis_ok = await self.redis.ping()
            redis_status = redis_ok
            _log.debug(
                "Redis health check: OK" if redis_ok else "Redis health check: FAILED"
            )
        except Exception as e:
            redis_status = False
            _log.error(f"Redis health check failed: {e}")

        # Calculate actual uptime
        current_time = datetime.now(UTC)
        uptime_seconds = int((current_time - settings.startup_time).total_seconds())

        response = SystemHealthResponse(
            database=database_status,
            redis=redis_status,
            uptime_seconds=uptime_seconds,
            version="1.0.0",
            environment=EnvironmentEnum.DEVELOPMENT,
            timestamp=current_time,
        )

        _log.info(
            f"System health check completed for admin {admin_user_id}: "
            f"db={database_status}, redis={redis_status}, uptime={uptime_seconds}s"
        )
        return response

    async def force_logout_user(
        self, admin_user_id: str | UUID, user_id: UUID
    ) -> ForceLogoutResponse:
        """Force logout a user.

        Args:
            admin_user_id: The admin user's ID (for admin check)
            user_id: The user's unique identifier
        Returns:
            ForceLogoutResponse: Force logout result
        """
        _log.warning(f"Admin {admin_user_id} attempting to force logout user {user_id}")

        await self._ensure_admin(admin_user_id)

        try:
            # Validate target user exists
            target_user = await self.db.get_user_by_id(str(user_id))
            if not target_user:
                _log.warning(f"Force logout failed: Target user {user_id} not found")
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="Target user not found.",
                )

            # Prevent admin from force logging out themselves
            if str(admin_user_id) == str(user_id):
                _log.warning(
                    f"Admin {admin_user_id} attempted to force logout themselves"
                )
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Cannot force logout yourself.",
                )

            # Invalidate all sessions for the user
            sessions_terminated = await self.redis.invalidate_user_sessions(
                str(user_id)
            )

            response = ForceLogoutResponse(
                user_id=user_id,
                sessions_terminated=sessions_terminated,
                timestamp=datetime.now(UTC),
                admin_user_id=UUID(str(admin_user_id)),
            )

            _log.warning(
                f"Force logout successful: Admin {admin_user_id} force logged out "
                f"user {user_id}, terminated {sessions_terminated} sessions"
            )
        except DatabaseError as e:
            _log.error(f"Database error: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except redis.ConnectionError as e:
            _log.error(f"Redis connection error: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="Redis service unavailable.",
            ) from e
        else:
            return response

    async def clear_redis_sessions(
        self, admin_user_id: str | UUID
    ) -> ClearSessionsResponse:
        """Clear all Redis sessions.

        Args:
            admin_user_id: The admin user's ID (for admin check)
        Returns:
            ClearSessionsResponse: Clear sessions result
        """
        _log.warning(f"Admin {admin_user_id} attempting to clear all Redis sessions")

        try:
            await self._ensure_admin(admin_user_id)

            # Get session stats before clearing
            stats = await self.redis.get_session_stats()
            sessions_to_clear = stats.total_sessions

            if sessions_to_clear == 0:
                _log.info("No sessions to clear")
                return ClearSessionsResponse(
                    sessions_cleared=0,
                    timestamp=datetime.now(UTC),
                    admin_user_id=UUID(str(admin_user_id)),
                    redis_keys_removed=0,
                )

            # Clear all sessions
            redis_keys_removed = await self.redis.clear_all_sessions()

            response = ClearSessionsResponse(
                sessions_cleared=sessions_to_clear,
                timestamp=datetime.now(UTC),
                admin_user_id=UUID(str(admin_user_id)),
                redis_keys_removed=redis_keys_removed,
            )

            _log.warning(
                f"All Redis sessions cleared by admin {admin_user_id}: "
                f"{sessions_to_clear} sessions, {redis_keys_removed} keys removed"
            )
        except redis.ConnectionError as e:
            _log.error(f"Redis connection error: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="Redis service unavailable.",
            ) from e
        else:
            return response
