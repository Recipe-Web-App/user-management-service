"""Routes package."""

from fastapi import APIRouter

from app.api.v1.routes import auth, health, notifications, social, users

api_router = APIRouter()
api_router.include_router(health.router)
api_router.include_router(auth.router)
api_router.include_router(users.router)
api_router.include_router(social.router)
api_router.include_router(notifications.router)
