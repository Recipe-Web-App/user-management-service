"""Main application entry point for the User Management Service."""

import uvicorn
from fastapi import FastAPI

app = FastAPI()


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
