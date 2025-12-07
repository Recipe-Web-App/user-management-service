"""Redis connection and session management."""

from collections.abc import AsyncGenerator

from redis.asyncio import Redis as AsyncRedis

from app.core.config import settings
from app.db.redis.redis_database_session import RedisDatabaseSession

# Global Redis connection pool
redis_client: AsyncRedis = AsyncRedis(
    host=settings.redis_host,
    port=settings.redis_port,
    password=settings.redis_password,
    db=settings.redis_db,
    decode_responses=True,
)


async def get_redis() -> AsyncGenerator[AsyncRedis, None]:
    """Get Redis client."""
    try:
        yield redis_client
    finally:
        # Redis handles connection cleanup automatically
        pass


async def get_redis_session() -> RedisDatabaseSession:
    """Get Redis database session."""
    return RedisDatabaseSession(redis_client)


async def init_redis() -> None:
    """Initialize Redis connection."""
    await redis_client.ping()  # type: ignore[misc]


async def close_redis() -> None:
    """Close Redis connection."""
    await redis_client.close()
