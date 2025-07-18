"""Custom SQL database session with common query methods."""

from datetime import UTC, datetime, timedelta
from http import HTTPStatus
from typing import Any

from sqlalchemy import func, select
from sqlalchemy.exc import DisconnectionError, OperationalError
from sqlalchemy.exc import TimeoutError as SQLTimeoutError
from sqlalchemy.ext.asyncio import AsyncSession

from app.core.logging import get_logger
from app.db.sql.models.user.user import User
from app.enums.user_role_enum import UserRoleEnum
from app.exceptions.custom_exceptions.database_exceptions import DatabaseError

_log = get_logger(__name__)


class SqlDatabaseSession(AsyncSession):
    """Custom SQL database session with common query methods."""

    def _classify_operational_error(self, error: OperationalError) -> DatabaseError:
        """Classify OperationalError based on PostgreSQL error codes.

        Args:
            error: The OperationalError to classify

        Returns:
            DatabaseError: Classified error with appropriate status code
        """
        error_code = (
            getattr(error.orig, "pgcode", None) if hasattr(error, "orig") else None
        )

        if error_code:
            _log.debug(f"PostgreSQL error code: {error_code}")

            # Connection and availability issues (503)
            if error_code in (
                DatabaseError.CONNECTION_EXCEPTION,
                DatabaseError.SQLCLIENT_UNABLE_TO_ESTABLISH_SQLCONNECTION,
                DatabaseError.CONNECTION_DOES_NOT_EXIST,
                DatabaseError.CONNECTION_FAILURE,
            ):
                return DatabaseError(
                    f"Database service unavailable: {error}",
                    status_code=HTTPStatus.SERVICE_UNAVAILABLE,
                )

            # Constraint violations (400)
            if error_code in (
                DatabaseError.UNIQUE_VIOLATION,
                DatabaseError.FOREIGN_KEY_VIOLATION,
                DatabaseError.CHECK_VIOLATION,
            ):
                return DatabaseError(
                    f"Data constraint violation: {error}",
                    status_code=HTTPStatus.BAD_REQUEST,
                )

            # Data type and format issues (422)
            if error_code in (
                DatabaseError.INVALID_TEXT_REPRESENTATION,
                DatabaseError.INVALID_BINARY_REPRESENTATION,
                DatabaseError.BAD_COPY_FILE_FORMAT,
                DatabaseError.UNTRANSLATABLE_CHARACTER,
            ):
                return DatabaseError(
                    f"Data format error: {error}",
                    status_code=HTTPStatus.UNPROCESSABLE_ENTITY,
                )

            # Permission and authorization issues (403)
            if error_code in (
                DatabaseError.INSUFFICIENT_PRIVILEGE,
                DatabaseError.DEPENDENT_PRIVILEGE_DESCRIPTORS_STILL_EXIST,
            ):
                return DatabaseError(
                    f"Database permission error: {error}",
                    status_code=HTTPStatus.FORBIDDEN,
                )

            # Resource exhaustion (503)
            if error_code in (
                DatabaseError.DISK_FULL,
                DatabaseError.OUT_OF_MEMORY,
                DatabaseError.TOO_MANY_CONNECTIONS,
            ):
                return DatabaseError(
                    f"Database resource exhausted: {error}",
                    status_code=HTTPStatus.SERVICE_UNAVAILABLE,
                )

        # Default to 500 for unknown operational errors
        return DatabaseError(
            f"Database operational error: {error}",
            status_code=HTTPStatus.INTERNAL_SERVER_ERROR,
        )

    async def _handle_database_error(self, operation: str, error: Exception) -> None:
        """Handle database errors with appropriate logging and re-raising.

        Args:
            operation: Description of the operation that failed
            error: The database error that occurred

        Raises:
            DatabaseError: Classified database error with appropriate status code
        """
        error_msg = f"Database {operation} failed: {error}"
        _log.error(error_msg)

        if isinstance(error, DisconnectionError | ConnectionRefusedError):
            raise DatabaseError(
                f"Database service unavailable: {operation} failed",
                status_code=HTTPStatus.SERVICE_UNAVAILABLE,
            ) from error
        if isinstance(error, SQLTimeoutError):
            raise DatabaseError(
                f"Database operation timed out: {operation}",
                status_code=HTTPStatus.SERVICE_UNAVAILABLE,
            ) from error
        if isinstance(error, OperationalError):
            raise self._classify_operational_error(error)

        # Re-raise the original error for other cases
        raise error

    async def execute(self, statement: Any, *args: Any, **kwargs: Any) -> Any:
        """Execute a database statement with error handling.

        Args:
            statement: SQLAlchemy statement to execute
            *args: Additional arguments
            **kwargs: Additional keyword arguments

        Returns:
            The result of the database operation

        Raises:
            DatabaseError: Classified database error with appropriate status code
        """
        try:
            return await super().execute(statement, *args, **kwargs)
        except Exception as e:
            await self._handle_database_error("execute", e)
            raise

    async def commit(self) -> None:
        """Commit the current transaction with error handling.

        Raises:
            DatabaseError: Classified database error with appropriate status code
        """
        try:
            await super().commit()
        except Exception as e:
            await self._handle_database_error("commit", e)
            raise

    async def refresh(self, instance: Any, *args: Any, **kwargs: Any) -> None:
        """Refresh an instance with error handling.

        Args:
            instance: The instance to refresh
            *args: Additional arguments
            **kwargs: Additional keyword arguments

        Raises:
            DatabaseError: Classified database error with appropriate status code
        """
        try:
            await super().refresh(instance, *args, **kwargs)
        except Exception as e:
            await self._handle_database_error("refresh", e)
            raise

    async def get_user_by_username(self, username: str) -> User | None:
        """Get user by username with error handling.

        Args:
            username: Username to search for

        Returns:
            User | None: User if found, None otherwise

        Raises:
            DatabaseError: Classified database error with appropriate status code
        """
        try:
            result = await self.execute(select(User).where(User.username == username))
            return result.scalar_one_or_none()
        except Exception as e:
            await self._handle_database_error("get_user_by_username", e)
            raise

    async def get_user_by_email(self, email: str) -> User | None:
        """Get user by email with error handling.

        Args:
            email: Email to search for

        Returns:
            User | None: User if found, None otherwise

        Raises:
            DatabaseError: Classified database error with appropriate status code
        """
        try:
            result = await self.execute(select(User).where(User.email == email))
            return result.scalar_one_or_none()
        except Exception as e:
            await self._handle_database_error("get_user_by_email", e)
            raise

    async def get_user_by_id(self, user_id: str) -> User | None:
        """Get user by ID with error handling.

        Args:
            user_id: User ID to search for

        Returns:
            User | None: User if found, None otherwise

        Raises:
            DatabaseError: Classified database error with appropriate status code
        """
        try:
            result = await self.execute(select(User).where(User.user_id == user_id))
            return result.scalar_one_or_none()
        except Exception as e:
            await self._handle_database_error("get_user_by_id", e)
            raise

    async def create_user(self, user: User) -> User:
        """Create a new user with error handling.

        Args:
            user: User to create

        Returns:
            User: Created user

        Raises:
            DatabaseError: Classified database error with appropriate status code
        """
        try:
            self.add(user)
            await self.commit()
            await self.refresh(user)
        except Exception as e:
            await self._handle_database_error("create_user", e)
            raise
        else:
            return user

    async def get_user_statistics(self) -> dict[str, int]:
        """Get comprehensive user statistics.

        Returns:
            dict[str, int]: Dictionary containing various user statistics

        Raises:
            DatabaseError: Classified database error with appropriate status code
        """
        try:
            # Total users
            total_result = await self.execute(select(func.count(User.user_id)))
            total_users = total_result.scalar() or 0

            # Active users
            active_result = await self.execute(
                select(func.count(User.user_id)).where(User.is_active.is_(True))
            )
            active_users = active_result.scalar() or 0

            # Deactivated users
            deactivated_result = await self.execute(
                select(func.count(User.user_id)).where(User.is_active.is_(False))
            )
            deactivated_users = deactivated_result.scalar() or 0

            # Admin users
            admin_result = await self.execute(
                select(func.count(User.user_id)).where(User.role == UserRoleEnum.ADMIN)
            )
            admin_users = admin_result.scalar() or 0

            # Recently registered users (last 30 days)
            thirty_days_ago = datetime.now(UTC) - timedelta(days=30)
            recent_result = await self.execute(
                select(func.count(User.user_id)).where(
                    User.created_at >= thirty_days_ago
                )
            )
            recently_registered = recent_result.scalar() or 0

            # Users with verified email (have email and are active)
            verified_result = await self.execute(
                select(func.count(User.user_id)).where(
                    User.email.isnot(None), User.is_active.is_(True)
                )
            )
            verified_users = verified_result.scalar() or 0
        except Exception as e:
            await self._handle_database_error("get_user_statistics", e)
            raise
        else:
            return {
                "total_users": total_users,
                "active_users": active_users,
                "deactivated_users": deactivated_users,
                "admin_users": admin_users,
                "recently_registered": recently_registered,
                "verified_users": verified_users,
            }

    async def get_users_online_count(self) -> int:
        """Get count of users currently online (with active sessions).

        Returns:
            int: Number of users currently online

        Raises:
            DatabaseError: Classified database error with appropriate status code
        """
        try:
            # This is a simplified implementation
            # In a real application, you would check active sessions in Redis
            # For now, we'll return active users as a proxy for online users
            result = await self.execute(
                select(func.count(User.user_id)).where(User.is_active.is_(True))
            )
            return result.scalar() or 0
        except Exception as e:
            await self._handle_database_error("get_users_online_count", e)
            raise

    async def get_user_retention_rate(self) -> float:
        """Calculate user retention rate based on active vs total users.

        Returns:
            float: User retention rate as percentage

        Raises:
            DatabaseError: Classified database error with appropriate status code
        """
        try:
            stats = await self.get_user_statistics()
            if stats["total_users"] == 0:
                return 0.0
            return round((stats["active_users"] / stats["total_users"]) * 100, 1)
        except Exception as e:
            await self._handle_database_error("get_user_retention_rate", e)
            raise

    async def get_average_registration_rate(self) -> float:
        """Calculate average daily user registrations over the last 30 days.

        Returns:
            float: Average daily registrations

        Raises:
            DatabaseError: Classified database error with appropriate status code
        """
        try:
            stats = await self.get_user_statistics()
            return round(stats["recently_registered"] / 30, 1)
        except Exception as e:
            await self._handle_database_error("get_average_registration_rate", e)
            raise
