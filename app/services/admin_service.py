"""Admin service for operational and privileged admin endpoints."""

from uuid import UUID

from fastapi import HTTPException, status

from app.api.v1.schemas.response.admin.redis_session_stats_response import (
    RedisSessionStatsResponse,
)
from app.api.v1.schemas.response.admin.user_stats_response import UserStatsResponse
from app.db.redis.redis_database_session import RedisDatabaseSession
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.enums.user_role_enum import UserRoleEnum


class AdminService:
    """Service for admin operations."""

    def __init__(self, db: SqlDatabaseSession, redis: RedisDatabaseSession) -> None:
        """Initialize admin service with database and redis sessions."""
        self.db = db
        self.redis = redis

    async def _ensure_admin(self, user_id: str | UUID) -> None:
        """Ensure the user is an admin, else raise 403."""
        user_id_str = str(user_id)
        user = await self.db.get_user_by_id(user_id_str)
        if not user or user.role != UserRoleEnum.ADMIN:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Admin privileges required.",
            )

    async def get_redis_session_stats(
        self, admin_user_id: str | UUID
    ) -> RedisSessionStatsResponse:
        """Return Redis session statistics using RedisDatabaseSession method.

        Args:
            admin_user_id: The admin user's ID (for admin check)
        Returns:
            RedisSessionStatsResponse: Redis session stats
        """
        await self._ensure_admin(admin_user_id)
        return await self.redis.get_session_stats()

    async def get_user_stats(self, admin_user_id: str | UUID) -> UserStatsResponse:
        """Return user statistics.

        Args:
            admin_user_id: The admin user's ID (for admin check)
        Returns:
            UserStatsResponse: User statistics
        """
        await self._ensure_admin(admin_user_id)

        # Get user statistics from database
        user_stats = await self.db.get_user_statistics()
        users_online = await self.db.get_users_online_count()
        retention_rate = await self.db.get_user_retention_rate()
        avg_registration_rate = await self.db.get_average_registration_rate()

        return UserStatsResponse(
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

    async def get_system_health(self, admin_user_id: str | UUID) -> dict:
        """Return system health status (dummy data).

        Args:
            admin_user_id: The admin user's ID (for admin check)
        Returns:
            dict: System health
        """
        await self._ensure_admin(admin_user_id)
        return {"database": "ok", "redis": "ok", "uptime_seconds": 123456}

    async def get_recent_logins(self, admin_user_id: str | UUID) -> dict:
        """Return recent user logins (dummy data).

        Args:
            admin_user_id: The admin user's ID (for admin check)
        Returns:
            dict: Recent logins
        """
        await self._ensure_admin(admin_user_id)
        return {
            "recent_logins": [
                {"user_id": "user1", "timestamp": "2025-07-17T01:00:00Z"},
                {"user_id": "user2", "timestamp": "2025-07-17T01:05:00Z"},
            ]
        }

    async def force_logout_user(self, admin_user_id: str | UUID, user_id: UUID) -> dict:
        """Force logout a user (dummy response).

        Args:
            admin_user_id: The admin user's ID (for admin check)
            user_id: The user's unique identifier
        Returns:
            dict: Force logout result
        """
        await self._ensure_admin(admin_user_id)
        return {"user_id": str(user_id), "status": "force-logout triggered (dummy)"}

    async def clear_redis_sessions(self, admin_user_id: str | UUID) -> dict:
        """Clear all Redis sessions (dummy response).

        Args:
            admin_user_id: The admin user's ID (for admin check)
        Returns:
            dict: Clear sessions result
        """
        await self._ensure_admin(admin_user_id)
        return {"status": "all sessions cleared (dummy)"}
