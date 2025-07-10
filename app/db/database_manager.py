"""Database connection and session management."""

from collections.abc import AsyncGenerator

from sqlalchemy.ext.asyncio import async_sessionmaker, create_async_engine

from app.core.config import settings
from app.db.models.base_database_model import BaseDatabaseModel
from app.db.session import DatabaseSession

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
    class_=DatabaseSession,
    expire_on_commit=False,
    autoflush=False,
    autocommit=False,
)


async def get_db() -> AsyncGenerator[DatabaseSession, None]:
    """Get database session."""
    async with AsyncSessionLocal() as session:
        try:
            yield session
        finally:
            await session.close()


async def init_db() -> None:
    """Initialize database tables."""
    async with engine.begin() as conn:
        await conn.run_sync(BaseDatabaseModel.metadata.create_all)
