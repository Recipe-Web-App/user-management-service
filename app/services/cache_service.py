"""Caching service for improved performance and reduced database load."""

import json
import logging
from typing import Any

from redis.asyncio import Redis

logger = logging.getLogger(__name__)


class CacheService:
    """Service for handling caching operations with Redis."""

    def __init__(self, redis_client: Redis):
        """Initialize cache service.

        Args:
            redis_client: Redis client instance
        """
        self.redis = redis_client
        self.default_ttl = 300  # 5 minutes default TTL

    async def get(self, key: str) -> Any | None:
        """Get value from cache.

        Args:
            key: Cache key

        Returns:
            Cached value if found, None otherwise
        """
        try:
            cached_data = await self.redis.get(key)
            if cached_data:
                logger.debug("Cache hit", extra={"key": key})
                return json.loads(cached_data)
        except Exception as e:
            logger.exception("Cache get error", extra={"key": key, "error": str(e)})
            return None
        else:
            logger.debug("Cache miss", extra={"key": key})
            return None

    async def set(self, key: str, value: Any, ttl: int | None = None) -> bool:
        """Set value in cache.

        Args:
            key: Cache key
            value: Value to cache
            ttl: Time to live in seconds (uses default if not provided)

        Returns:
            True if successful, False otherwise
        """
        try:
            ttl = ttl or self.default_ttl
            serialized_value = json.dumps(value, default=str)

            result = await self.redis.setex(key, ttl, serialized_value)

            if result:
                logger.debug("Cache set successful", extra={"key": key, "ttl": ttl})

            return bool(result)
        except Exception as e:
            logger.exception("Cache set error", extra={"key": key, "error": str(e)})
            return False

    async def delete(self, key: str) -> bool:
        """Delete value from cache.

        Args:
            key: Cache key to delete

        Returns:
            True if key was deleted, False otherwise
        """
        try:
            result = await self.redis.delete(key)

            if result:
                logger.debug("Cache delete successful", extra={"key": key})

            return bool(result)
        except Exception as e:
            logger.exception("Cache delete error", extra={"key": key, "error": str(e)})
            return False

    async def exists(self, key: str) -> bool:
        """Check if key exists in cache.

        Args:
            key: Cache key to check

        Returns:
            True if key exists, False otherwise
        """
        try:
            result = await self.redis.exists(key)
            return bool(result)

        except Exception as e:
            logger.exception(
                "Cache exists check error", extra={"key": key, "error": str(e)}
            )
            return False

    async def clear_pattern(self, pattern: str) -> int:
        """Clear all cache keys matching a pattern.

        Args:
            pattern: Redis pattern (e.g., "user:*")

        Returns:
            Number of keys deleted
        """
        try:
            keys = await self.redis.keys(pattern)
            if keys:
                deleted_count = await self.redis.delete(*keys)
                logger.info(
                    "Cache pattern cleared",
                    extra={"pattern": pattern, "deleted_count": deleted_count},
                )
                return deleted_count
        except Exception as e:
            logger.exception(
                "Cache pattern clear error", extra={"pattern": pattern, "error": str(e)}
            )
            return 0
        else:
            return 0

    async def get_cache_stats(self) -> dict[str, Any]:
        """Get cache statistics.

        Returns:
            Dictionary with cache statistics
        """
        try:
            info = await self.redis.info("memory")
            keyspace_info = await self.redis.info("keyspace")

            stats: dict[str, Any] = {
                "memory_usage_bytes": info.get("used_memory", 0),
                "memory_usage_human": info.get("used_memory_human", "0B"),
                "connected_clients": info.get("connected_clients", 0),
                "keyspace_hits": info.get("keyspace_hits", 0),
                "keyspace_misses": info.get("keyspace_misses", 0),
                "total_keys": 0,
            }

            # Calculate total keys across all databases
            for db_name, db_info in keyspace_info.items():
                if db_name.startswith("db"):
                    keys_count = int(db_info.split(",")[0].split("=")[1])
                    stats["total_keys"] += keys_count

            # Calculate hit rate
            hits = stats["keyspace_hits"]
            misses = stats["keyspace_misses"]
            total_requests = hits + misses

            if total_requests > 0:
                stats["hit_rate"] = round((hits / total_requests) * 100, 2)
            else:
                stats["hit_rate"] = 0.0
        except Exception as e:
            logger.exception("Cache stats error", extra={"error": str(e)})
            return {
                "error": "Unable to retrieve cache statistics",
                "memory_usage_bytes": 0,
                "total_keys": 0,
                "hit_rate": 0.0,
            }
        else:
            return stats


class CacheKeyBuilder:
    """Helper class for building consistent cache keys."""

    @staticmethod
    def user_profile(user_id: str) -> str:
        """Build cache key for user profile."""
        return f"user:profile:{user_id}"

    @staticmethod
    def user_followers(user_id: str, offset: int = 0, limit: int = 20) -> str:
        """Build cache key for user followers."""
        return f"user:followers:{user_id}:{offset}:{limit}"

    @staticmethod
    def user_following(user_id: str, offset: int = 0, limit: int = 20) -> str:
        """Build cache key for user following."""
        return f"user:following:{user_id}:{offset}:{limit}"

    @staticmethod
    def user_notifications(user_id: str, offset: int = 0, limit: int = 20) -> str:
        """Build cache key for user notifications."""
        return f"user:notifications:{user_id}:{offset}:{limit}"

    @staticmethod
    def user_search(query: str, offset: int = 0, limit: int = 20) -> str:
        """Build cache key for user search results."""
        return f"search:users:{query}:{offset}:{limit}"

    @staticmethod
    def system_health() -> str:
        """Build cache key for system health."""
        return "system:health"

    @staticmethod
    def user_session(session_id: str) -> str:
        """Build cache key for user session."""
        return f"session:{session_id}"


class CacheTTL:
    """Cache TTL constants for different data types."""

    USER_PROFILE = 600  # 10 minutes
    USER_FOLLOWERS = 300  # 5 minutes
    USER_FOLLOWING = 300  # 5 minutes
    USER_NOTIFICATIONS = 60  # 1 minute
    USER_SEARCH = 180  # 3 minutes
    SYSTEM_HEALTH = 30  # 30 seconds
    USER_SESSION = 1800  # 30 minutes
