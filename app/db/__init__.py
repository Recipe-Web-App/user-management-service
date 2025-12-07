"""Database package exports."""

from .redis.redis_database_manager import get_redis, get_redis_session, init_redis
from .sql.sql_database_manager import get_db, init_db

# For backward compatibility and unified access
__all__ = ["get_db", "get_redis", "get_redis_session", "init_db", "init_redis"]
