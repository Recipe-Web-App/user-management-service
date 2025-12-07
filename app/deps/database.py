"""Database dependency providers."""

from collections.abc import AsyncGenerator
from typing import Annotated

from fastapi import Depends
from redis.asyncio import Redis

from app.db.redis.redis_database_manager import get_redis, get_redis_session
from app.db.redis.redis_database_session import RedisDatabaseSession
from app.db.sql.sql_database_manager import get_db
from app.db.sql.sql_database_session import SqlDatabaseSession


async def get_db_session() -> AsyncGenerator[SqlDatabaseSession, None]:
    """Get SQL database session for dependency injection.

    Creates and yields an async SQL database session that can be used
    for database operations in FastAPI endpoints. The session is
    automatically managed and closed after use.

    Returns:
        AsyncGenerator[SqlDatabaseSession, None]: Async generator that yields
            a SQL database session instance.

    Yields:
        SqlDatabaseSession: Database session for performing SQL operations.

    Example:
        ```python
        async def some_endpoint(db: DatabaseSession):
            user = await db.get_user_by_id("123")
        ```
    """
    async for session in get_db():
        yield session


async def get_redis_client() -> AsyncGenerator[Redis, None]:
    """Get Redis client for dependency injection.

    Creates and yields an async Redis client that can be used for
    caching and session management operations in FastAPI endpoints.
    The client connection is automatically managed.

    Returns:
        AsyncGenerator[Redis, None]: Async generator that yields
            a Redis client instance.

    Yields:
        Redis: Redis client for performing cache and session operations.

    Example:
        ```python
        async def some_endpoint(redis: RedisSession):
            await redis.set("key", "value")
        ```
    """

    async for redis_client in get_redis():
        yield redis_client


async def get_redis_db_session() -> AsyncGenerator[RedisDatabaseSession, None]:
    """Get Redis database session for dependency injection.

    Creates and yields a Redis database session that provides
    higher-level database operations on top of the Redis client.
    Useful for structured data operations and session management.

    Returns:
        AsyncGenerator[RedisDatabaseSession, None]: Async generator that yields
            a Redis database session instance.

    Yields:
        RedisDatabaseSession: Redis database session for structured operations.

    Example:
        ```python
        async def some_endpoint(redis_db: RedisDbSession):
            await redis_db.store_user_session(user_id, session_data)
        ```
    """
    redis_session = await get_redis_session()
    yield redis_session


# Type aliases for dependency injection
DatabaseSession = Annotated[SqlDatabaseSession, Depends(get_db_session)]
RedisSession = Annotated[Redis, Depends(get_redis_client)]
RedisDbSession = Annotated[RedisDatabaseSession, Depends(get_redis_db_session)]
