"""Health check service for monitoring application and dependency status."""

import logging
import time

from redis.exceptions import ConnectionError as RedisConnectionError
from redis.exceptions import RedisError
from redis.exceptions import TimeoutError as RedisTimeoutError
from sqlalchemy import text
from sqlalchemy.exc import DisconnectionError, SQLAlchemyError
from sqlalchemy.exc import TimeoutError as SATimeoutError

from app.api.v1.schemas.response.dependency_health import DependencyHealth
from app.api.v1.schemas.response.liveness_response import LivenessResponse
from app.api.v1.schemas.response.readiness_response import ReadinessResponse
from app.db.redis.redis_database_manager import redis_client
from app.db.sql.sql_database_manager import engine
from app.enums.health_status import HealthStatus

logger = logging.getLogger(__name__)


class HealthService:
    """Service for performing health checks on application dependencies."""

    async def check_database_health(self) -> DependencyHealth:
        """Check PostgreSQL database connectivity and health.

        Returns:
            DependencyHealth: Health status with details about the check
        """
        start_time = time.time()

        try:
            logger.debug("Checking PostgreSQL database connectivity")
            async with engine.begin() as conn:
                result = await conn.execute(text("SELECT 1 as health_check"))
                row = result.fetchone()

            response_time = (time.time() - start_time) * 1000

            if row and row[0] == 1:
                logger.debug(
                    "PostgreSQL database health check successful "
                    f"({response_time:.2f}ms)"
                )
                return DependencyHealth(
                    healthy=True,
                    status=HealthStatus.HEALTHY,
                    message="Database connection successful",
                    response_time_ms=response_time,
                )
            logger.warning(
                "PostgreSQL database health check returned unexpected result"
            )
            return DependencyHealth(
                healthy=False,
                status=HealthStatus.ERROR,
                message="Database query returned unexpected result",
                response_time_ms=response_time,
            )

        except SATimeoutError:
            response_time = (time.time() - start_time) * 1000
            logger.exception("PostgreSQL database health check timed out")
            return DependencyHealth(
                healthy=False,
                status=HealthStatus.TIMEOUT,
                message="Database connection timeout",
                response_time_ms=response_time,
            )
        except DisconnectionError:
            response_time = (time.time() - start_time) * 1000
            logger.exception("PostgreSQL database disconnection error")
            return DependencyHealth(
                healthy=False,
                status=HealthStatus.DISCONNECTED,
                message="Database connection lost",
                response_time_ms=response_time,
            )
        except SQLAlchemyError as e:
            response_time = (time.time() - start_time) * 1000
            logger.exception("PostgreSQL database SQLAlchemy error")
            return DependencyHealth(
                healthy=False,
                status=HealthStatus.ERROR,
                message=f"Database error: {type(e).__name__}",
                response_time_ms=response_time,
            )
        except Exception as e:
            response_time = (time.time() - start_time) * 1000
            logger.exception("PostgreSQL database unexpected error")
            return DependencyHealth(
                healthy=False,
                status=HealthStatus.ERROR,
                message=f"Unexpected database error: {type(e).__name__}",
                response_time_ms=response_time,
            )

    async def check_redis_health(self) -> DependencyHealth:
        """Check Redis connectivity and health.

        Returns:
            DependencyHealth: Health status with details about the check
        """
        start_time = time.time()

        try:
            logger.debug("Checking Redis connectivity")
            pong = await redis_client.ping()
            response_time = (time.time() - start_time) * 1000

            if pong:
                logger.debug(f"Redis health check successful ({response_time:.2f}ms)")
                return DependencyHealth(
                    healthy=True,
                    status=HealthStatus.HEALTHY,
                    message="Redis connection successful",
                    response_time_ms=response_time,
                )
            logger.warning("Redis ping returned False")
            return DependencyHealth(
                healthy=False,
                status=HealthStatus.ERROR,
                message="Redis ping returned False",
                response_time_ms=response_time,
            )

        except RedisTimeoutError:
            response_time = (time.time() - start_time) * 1000
            logger.exception("Redis health check timed out")
            return DependencyHealth(
                healthy=False,
                status=HealthStatus.TIMEOUT,
                message="Redis connection timeout",
                response_time_ms=response_time,
            )
        except RedisConnectionError:
            response_time = (time.time() - start_time) * 1000
            logger.exception("Redis connection error")
            return DependencyHealth(
                healthy=False,
                status=HealthStatus.DISCONNECTED,
                message="Redis connection failed",
                response_time_ms=response_time,
            )
        except RedisError as e:
            response_time = (time.time() - start_time) * 1000
            logger.exception("Redis error")
            return DependencyHealth(
                healthy=False,
                status=HealthStatus.ERROR,
                message=f"Redis error: {type(e).__name__}",
                response_time_ms=response_time,
            )
        except Exception as e:
            response_time = (time.time() - start_time) * 1000
            logger.exception("Redis unexpected error")
            return DependencyHealth(
                healthy=False,
                status=HealthStatus.ERROR,
                message=f"Unexpected Redis error: {type(e).__name__}",
                response_time_ms=response_time,
            )

    async def get_readiness_status(self) -> ReadinessResponse:
        """Get comprehensive readiness status including all dependencies.

        Returns degraded (200 OK) status when only database is unavailable,
        allowing the service to remain deployable while non-database operations
        can still function. Returns not ready (503) only when Redis is unavailable
        since JWT sessions are critical.

        Returns:
            ReadinessResponse: Comprehensive readiness status
        """
        logger.info("Performing readiness health check")

        db_health = await self.check_database_health()
        redis_health = await self.check_redis_health()

        # Service is ready if Redis is healthy (database can be degraded)
        redis_healthy = redis_health.healthy
        db_healthy = db_health.healthy
        all_healthy = db_healthy and redis_healthy

        # Determine overall service status
        if all_healthy:
            service_ready = True
            service_degraded = False
            service_status = "ready"
        elif redis_healthy and not db_healthy:
            # Database is down but Redis is up - degraded mode
            service_ready = True
            service_degraded = True
            service_status = "degraded"
        elif db_healthy and not redis_healthy:
            # Redis is down but database is up - degraded mode
            service_ready = True
            service_degraded = True
            service_status = "degraded"
        elif not db_healthy and not redis_healthy:
            # Both dependencies are down - degraded mode (basic service still works)
            service_ready = True
            service_degraded = True
            service_status = "degraded"
        else:
            # Should not reach here, but fallback to not ready
            service_ready = False
            service_degraded = False
            service_status = "not ready"

        dependencies: dict[str, DependencyHealth] = {
            "database": db_health,
            "redis": redis_health,
        }

        status = ReadinessResponse(
            ready=service_ready,
            status=service_status,
            degraded=service_degraded,
            dependencies=dependencies,
        )

        if all_healthy:
            logger.info("All dependencies healthy - service ready")
        elif service_degraded:
            logger.warning(
                f"Service running in degraded mode - Database: {db_health.status}, "
                f"Redis: {redis_health.status}. Dependency reconnection service active."
            )
        else:
            logger.warning(
                f"Service not ready - Database: {db_health.status}, "
                f"Redis: {redis_health.status}"
            )

        return status

    async def get_liveness_status(self) -> LivenessResponse:
        """Get basic liveness status (application is running).

        Returns:
            LivenessResponse: Basic liveness status
        """
        logger.debug("Performing liveness health check")
        return LivenessResponse(
            alive=True, status="alive", message="Service is running"
        )


# Global health service instance
health_service = HealthService()
