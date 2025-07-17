"""Admin service for operational and privileged admin endpoints."""

from uuid import UUID

from fastapi import HTTPException, status

from app.api.v1.schemas.response.admin.redis_session_stats_response import (
    RedisSessionStatsResponse,
)
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

    async def get_user_stats(self, admin_user_id: str | UUID) -> dict:
        """Return user statistics (dummy data).

        Args:
            admin_user_id: The admin user's ID (for admin check)
        Returns:
            dict: User stats
        """
        await self._ensure_admin(admin_user_id)
        return {
            "total_users": 10000,
            "active_users": 9000,
            "recently_registered": 50,
            "deactivated_users": 100,
        }

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
