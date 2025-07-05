"""Main application entry point for the User Management Service."""

import uvicorn
from fastapi import FastAPI

from app.api.v1.routes import api_router

app = FastAPI(
    title="User Management Service",
    version="1.0.0",
    description="API for managing users",
)
app.include_router(api_router, prefix="/api/v1")


def main() -> None:
    """Run the main application function."""
    uvicorn.run(
        "app.main:app",
        host="127.0.0.1",
        port=8000,
        reload=True,
    )
    uvicorn.run(
        "app.main:app",
        host="127.0.0.1",
        port=8000,
        reload=True,
    )


if __name__ == "__main__":
    main()
