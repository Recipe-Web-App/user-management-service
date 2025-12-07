"""SQL database connection and session management."""

from collections.abc import AsyncGenerator

from sqlalchemy.ext.asyncio import async_sessionmaker, create_async_engine

from app.core.config import settings
from app.db.sql.models.base_sql_model import BaseSqlModel
from app.db.sql.sql_database_session import SqlDatabaseSession

DATABASE_URL = (
    f"postgresql+asyncpg://{settings.user_management_db_user}:"
    f"{settings.user_management_db_password}@"
    f"{settings.postgres_host}:{settings.postgres_port}/"
    f"{settings.postgres_db}"
)

engine = create_async_engine(
    DATABASE_URL,
    echo=False,
    pool_pre_ping=True,
    pool_recycle=300,
    pool_size=10,
    max_overflow=20,
    pool_timeout=30,
)

AsyncSessionLocal = async_sessionmaker(
    engine,
    class_=SqlDatabaseSession,
    expire_on_commit=False,
    autoflush=False,
    autocommit=False,
)


async def get_db() -> AsyncGenerator[SqlDatabaseSession, None]:
    """Get database session for dependency injection.

    Creates and yields an async database session that can be used
    for database operations. The session is automatically closed
    after use, ensuring proper resource cleanup.

    Returns:
        AsyncGenerator[SqlDatabaseSession, None]: Async generator that yields
            a database session instance.

    Yields:
        SqlDatabaseSession: Database session for performing operations.

    Example:
        ```python
        async def some_endpoint(db: SqlDatabaseSession = Depends(get_db)):
            user = await db.get_user_by_id("123")
        ```
    """
    async with AsyncSessionLocal() as session:
        try:
            yield session
        finally:
            await session.close()


async def init_db() -> None:
    """Initialize the database by creating all tables.

    Creates all database tables defined in the SQLAlchemy models
    if they don't already exist. This function should be called
    during application startup to ensure the database schema
    is properly set up.

    Note:
        This function uses `create_all()` which only creates tables
        that don't exist. Existing tables are not modified.

    Raises:
        DatabaseError: If database connection fails or table creation fails.
    """

    async with engine.begin() as conn:
        await conn.run_sync(BaseSqlModel.metadata.create_all)
