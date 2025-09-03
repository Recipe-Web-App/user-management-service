"""Routes package."""

from fastapi import APIRouter

from app.api.v1.routes import admin, health, metrics, notifications, social, users

api_router = APIRouter()
api_router.include_router(admin.router)
api_router.include_router(health.router)
api_router.include_router(metrics.router)
api_router.include_router(users.router)
api_router.include_router(social.router)
api_router.include_router(notifications.router)

__all__ = ["api_router"]
