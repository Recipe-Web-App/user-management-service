"""Main application entry point for the User Management Service."""

import contextlib
import logging
from collections.abc import AsyncGenerator

import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.api.v1.routes import api_router
from app.core.config import settings
from app.core.logging import configure_logging
from app.db.database_reconnection import database_reconnection_service
from app.db.redis.redis_database_manager import init_redis
from app.db.sql.sql_database_manager import init_db
from app.exceptions.handlers import unhandled_exception_handler
from app.middleware.performance_middleware import PerformanceMiddleware
from app.middleware.request_id_middleware import request_id_middleware
from app.middleware.security_middleware import SecurityHeadersMiddleware

_log = logging.getLogger(__name__)


@contextlib.asynccontextmanager
async def lifespan(_: FastAPI) -> AsyncGenerator[None, None]:
    """Configure logging and database when the application starts."""
    configure_logging()

    try:
        await init_db()
    except Exception as e:
        _log.critical("Could not initialize sql database: %s", e)

    try:
        await init_redis()
    except Exception as e:
        _log.critical("Could not initialize redis database: %s", e)

    # Start database reconnection service
    try:
        await database_reconnection_service.start()
        _log.info("Database reconnection service started")
    except Exception as e:
        _log.error("Could not start database reconnection service: %s", e)

    yield

    # Stop database reconnection service
    try:
        await database_reconnection_service.stop()
        _log.info("Database reconnection service stopped")
    except Exception as e:
        _log.error("Error stopping database reconnection service: %s", e)


app = FastAPI(
    title="User Management Service",
    version="1.0.0",
    description="API for managing users",
    lifespan=lifespan,
)

# Add performance monitoring middleware
app.add_middleware(PerformanceMiddleware, enable_metrics=True)

# Add security headers middleware
app.add_middleware(SecurityHeadersMiddleware)

# Configure CORS
allowed_origins = settings.allowed_origin_hosts.split(",")
app.add_middleware(
    CORSMiddleware,
    allow_origins=allowed_origins,
    allow_credentials=settings.allowed_credentials,
    allow_methods=["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"],
    allow_headers=["*"],
)

app.add_exception_handler(Exception, unhandled_exception_handler)
app.middleware("http")(request_id_middleware)
app.include_router(api_router, prefix="/api/v1")


def main() -> None:
    """Run the main application function."""
    uvicorn.run(
        "app.main:app",
        host="127.0.0.1",
        port=8000,
        reload=True,
    )


if __name__ == "__main__":
    main()
