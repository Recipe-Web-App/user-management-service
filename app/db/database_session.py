"""Custom database session with common query methods."""

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.models.user.user import User


class DatabaseSession(AsyncSession):
    """Custom database session with common query methods."""

    async def get_user_by_username(self, username: str) -> User | None:
        """Get user by username."""
        result = await self.execute(select(User).where(User.username == username))
        return result.scalar_one_or_none()

    async def get_user_by_email(self, email: str) -> User | None:
        """Get user by email."""
        result = await self.execute(select(User).where(User.email == email))
        return result.scalar_one_or_none()

    async def create_user(self, user: User) -> User:
        """Create a new user."""
        self.add(user)
        await self.commit()
        await self.refresh(user)
        return user
