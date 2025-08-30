"""Database reconnection service for handling database and Redis failures."""

import asyncio
import contextlib
import logging
import time

from redis.exceptions import ConnectionError as RedisConnectionError
from redis.exceptions import RedisError
from redis.exceptions import TimeoutError as RedisTimeoutError
from sqlalchemy import text
from sqlalchemy.exc import SQLAlchemyError

from app.db.redis.redis_database_manager import redis_client
from app.db.sql.sql_database_manager import engine

logger = logging.getLogger(__name__)


class DatabaseReconnectionService:
    """Service for managing database and Redis reconnection attempts."""

    def __init__(
        self,
        reconnect_interval_seconds: int = 30,
        max_consecutive_failures: int = 10,
    ) -> None:
        """Initialize the database reconnection service.

        Args:
            reconnect_interval_seconds: Interval between reconnection attempts
            max_consecutive_failures: Maximum consecutive failures before
                extending the reconnection interval
        """
        self.reconnect_interval_seconds = reconnect_interval_seconds
        self.max_consecutive_failures = max_consecutive_failures
        self._task: asyncio.Task[None] | None = None
        self._shutdown_event = asyncio.Event()

        # Database connection state
        self._db_connection_healthy = False
        self._db_consecutive_failures = 0
        self._db_last_success_time: float | None = None
        self._db_last_failure_time: float | None = None

        # Redis connection state
        self._redis_connection_healthy = False
        self._redis_consecutive_failures = 0
        self._redis_last_success_time: float | None = None
        self._redis_last_failure_time: float | None = None

    @property
    def is_db_connection_healthy(self) -> bool:
        """Check if the database connection is currently healthy.

        Returns:
            bool: True if connection is healthy, False otherwise
        """
        return self._db_connection_healthy

    @property
    def is_redis_connection_healthy(self) -> bool:
        """Check if the Redis connection is currently healthy.

        Returns:
            bool: True if connection is healthy, False otherwise
        """
        return self._redis_connection_healthy

    @property
    def db_last_success_time(self) -> float | None:
        """Get the timestamp of the last successful database connection.

        Returns:
            Optional[float]: Timestamp of last success, None if never successful
        """
        return self._db_last_success_time

    @property
    def redis_last_success_time(self) -> float | None:
        """Get the timestamp of the last successful Redis connection.

        Returns:
            Optional[float]: Timestamp of last success, None if never successful
        """
        return self._redis_last_success_time

    @property
    def db_consecutive_failures(self) -> int:
        """Get the number of consecutive database connection failures.

        Returns:
            int: Number of consecutive failures
        """
        return self._db_consecutive_failures

    @property
    def redis_consecutive_failures(self) -> int:
        """Get the number of consecutive Redis connection failures.

        Returns:
            int: Number of consecutive failures
        """
        return self._redis_consecutive_failures

    async def _check_database_connection(self) -> bool:
        """Check if the database connection is working.

        Returns:
            bool: True if connection is healthy, False otherwise
        """
        try:
            async with engine.begin() as conn:
                result = await conn.execute(text("SELECT 1"))
                row = result.fetchone()
                return row is not None and row[0] == 1
        except SQLAlchemyError:
            return False
        except Exception:
            logger.exception("Unexpected error during database connection check")
            return False

    async def _check_redis_connection(self) -> bool:
        """Check if the Redis connection is working.

        Returns:
            bool: True if connection is healthy, False otherwise
        """
        try:
            pong = await redis_client.ping()
        except (RedisConnectionError, RedisTimeoutError, RedisError):
            return False
        except Exception:
            logger.exception("Unexpected error during Redis connection check")
            return False
        else:
            return pong is True

    async def _attempt_database_reconnection(self) -> bool:
        """Attempt to reconnect to the database.

        Returns:
            bool: True if reconnection successful, False otherwise
        """
        logger.info("Attempting database reconnection")

        try:
            # Dispose existing connections to force new connections
            await engine.dispose()

            # Test the new connection
            if await self._check_database_connection():
                self._db_connection_healthy = True
                self._db_consecutive_failures = 0
                self._db_last_success_time = time.time()
                logger.info("Database reconnection successful")
                return True

            self._db_connection_healthy = False
            self._db_consecutive_failures += 1
            self._db_last_failure_time = time.time()
            logger.warning(
                "Database reconnection failed "
                f"({self._db_consecutive_failures} consecutive failures)"
            )
        except Exception:
            self._db_connection_healthy = False
            self._db_consecutive_failures += 1
            self._db_last_failure_time = time.time()
            logger.exception(
                "Database reconnection attempt failed "
                f"({self._db_consecutive_failures} consecutive failures)"
            )

        return False

    async def _attempt_redis_reconnection(self) -> bool:
        """Attempt to reconnect to Redis.

        Returns:
            bool: True if reconnection successful, False otherwise
        """
        logger.info("Attempting Redis reconnection")

        try:
            # Test the Redis connection (Redis client auto-reconnects)
            if await self._check_redis_connection():
                self._redis_connection_healthy = True
                self._redis_consecutive_failures = 0
                self._redis_last_success_time = time.time()
                logger.info("Redis reconnection successful")
                return True

            self._redis_connection_healthy = False
            self._redis_consecutive_failures += 1
            self._redis_last_failure_time = time.time()
            logger.warning(
                "Redis reconnection failed "
                f"({self._redis_consecutive_failures} consecutive failures)"
            )
        except Exception:
            self._redis_connection_healthy = False
            self._redis_consecutive_failures += 1
            self._redis_last_failure_time = time.time()
            logger.exception(
                "Redis reconnection attempt failed "
                f"({self._redis_consecutive_failures} consecutive failures)"
            )

        return False

    def _calculate_reconnect_interval(self) -> int:
        """Calculate the next reconnection interval based on failure count.

        Returns:
            int: Reconnection interval in seconds
        """
        # Use the maximum failures from either dependency for interval calculation
        max_failures = max(
            self._db_consecutive_failures, self._redis_consecutive_failures
        )

        if max_failures <= self.max_consecutive_failures:
            return self.reconnect_interval_seconds

        # Exponential backoff with cap at 300 seconds (5 minutes)
        backoff_multiplier = min(
            2 ** (max_failures - self.max_consecutive_failures), 10
        )
        return min(self.reconnect_interval_seconds * backoff_multiplier, 300)

    async def _reconnection_loop(self) -> None:
        """Main reconnection loop that runs in the background."""
        logger.info("Starting database reconnection service")

        while not self._shutdown_event.is_set():
            try:
                # Check database connection
                db_healthy = await self._check_database_connection()
                if db_healthy:
                    if not self._db_connection_healthy:
                        # Database was previously unhealthy but is now healthy
                        self._db_connection_healthy = True
                        self._db_consecutive_failures = 0
                        self._db_last_success_time = time.time()
                        logger.info("Database connection restored")
                else:
                    # Database is unhealthy, attempt reconnection
                    if self._db_connection_healthy:
                        # Database was previously healthy but is now unhealthy
                        self._db_connection_healthy = False
                        self._db_consecutive_failures = 1
                        self._db_last_failure_time = time.time()
                        logger.warning(
                            "Database connection lost, starting reconnection attempts"
                        )
                    await self._attempt_database_reconnection()

                # Check Redis connection
                redis_healthy = await self._check_redis_connection()
                if redis_healthy:
                    if not self._redis_connection_healthy:
                        # Redis was previously unhealthy but is now healthy
                        self._redis_connection_healthy = True
                        self._redis_consecutive_failures = 0
                        self._redis_last_success_time = time.time()
                        logger.info("Redis connection restored")
                else:
                    # Redis is unhealthy, attempt reconnection
                    if self._redis_connection_healthy:
                        # Redis was previously healthy but is now unhealthy
                        self._redis_connection_healthy = False
                        self._redis_consecutive_failures = 1
                        self._redis_last_failure_time = time.time()
                        logger.warning(
                            "Redis connection lost, starting reconnection attempts"
                        )
                    await self._attempt_redis_reconnection()

                # Wait before next check/reconnection attempt
                interval = self._calculate_reconnect_interval()
                with contextlib.suppress(TimeoutError):
                    await asyncio.wait_for(
                        self._shutdown_event.wait(), timeout=interval
                    )
                    # If we reach here, shutdown was requested
                    break
                # Timeout is expected, continue the loop

            except Exception:
                logger.exception("Error in database reconnection loop")
                # Wait a bit before retrying to avoid tight error loops
                with contextlib.suppress(TimeoutError):
                    await asyncio.wait_for(self._shutdown_event.wait(), timeout=5)
                    # If we reach here, shutdown was requested
                    break
                # Timeout is expected, continue the loop

        logger.info("Database reconnection service stopped")

    async def start(self) -> None:
        """Start the database reconnection service."""
        if self._task is not None and not self._task.done():
            logger.warning("Database reconnection service is already running")
            return

        self._shutdown_event.clear()
        self._task = asyncio.create_task(self._reconnection_loop())
        logger.info("Database reconnection service started")

    async def stop(self) -> None:
        """Stop the database reconnection service."""
        if self._task is None:
            return

        logger.info("Stopping database reconnection service")
        self._shutdown_event.set()

        if not self._task.done():
            try:
                await asyncio.wait_for(self._task, timeout=5.0)
            except TimeoutError:
                logger.warning(
                    "Database reconnection service did not stop gracefully, "
                    "cancelling"
                )
                self._task.cancel()
                with contextlib.suppress(asyncio.CancelledError):
                    await self._task

        self._task = None
        logger.info("Database reconnection service stopped")


# Global database reconnection service instance
database_reconnection_service = DatabaseReconnectionService()
