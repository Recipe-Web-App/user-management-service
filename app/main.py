"""Main application entry point for the User Management Service."""

import contextlib
import logging
from collections.abc import AsyncGenerator

import uvicorn
from fastapi import FastAPI

from app.api.v1.routes import api_router
from app.core.logging import configure_logging
from app.db.database_manager import init_db
from app.exceptions.handlers import unhandled_exception_handler
from app.middleware.request_id_middleware import request_id_middleware

_log = logging.getLogger(__name__)


@contextlib.asynccontextmanager
async def lifespan(_: FastAPI) -> AsyncGenerator[None, None]:
    """Configure logging and database when the application starts."""
    configure_logging()

    try:
        await init_db()
    except Exception as e:
        _log.critical("Could not initialize database: %s", e)

    yield


app = FastAPI(
    title="User Management Service",
    version="1.0.0",
    description="API for managing users",
    lifespan=lifespan,
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
