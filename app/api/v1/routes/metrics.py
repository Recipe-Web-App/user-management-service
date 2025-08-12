"""Metrics endpoints for performance monitoring and observability."""

import asyncio
import logging
import os
from datetime import UTC, datetime
from typing import Any

import psutil
from fastapi import APIRouter

from app.deps.auth import CurrentAdminUser
from app.deps.database import RedisSession
from app.services.cache_service import CacheService

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/metrics", tags=["metrics"])


@router.get("/performance")
async def get_performance_metrics(
    current_admin: CurrentAdminUser,
) -> dict[str, Any]:
    """Get application performance metrics.

    Requires admin authentication.

    Returns:
        Performance metrics including response times and request counts
    """
    # Note: In a real implementation, you would get these from a metrics store
    # For now, we'll return placeholder data that could be populated by
    # the PerformanceMiddleware or an APM system

    metrics: dict[str, Any] = {
        "response_times": {
            "average_ms": 45.2,
            "p50_ms": 35.0,
            "p95_ms": 120.0,
            "p99_ms": 250.0,
        },
        "request_counts": {
            "total_requests": 15420,
            "requests_per_minute": 125,
            "active_sessions": 89,
        },
        "error_rates": {
            "total_errors": 23,
            "error_rate_percent": 0.15,
            "4xx_errors": 18,
            "5xx_errors": 5,
        },
        "database": {
            "active_connections": 8,
            "max_connections": 20,
            "avg_query_time_ms": 12.3,
            "slow_queries_count": 2,
        },
    }

    logger.info(
        "Performance metrics retrieved",
        extra={"admin_id": str(current_admin.user_id)},
    )
    return metrics


@router.get("/cache")
async def get_cache_metrics(
    current_admin: CurrentAdminUser,
    redis_session: RedisSession,
) -> dict[str, Any]:
    """Get cache performance metrics.

    Requires admin authentication.

    Returns:
        Cache statistics and performance metrics
    """
    cache_service = CacheService(redis_session)
    cache_stats = await cache_service.get_cache_stats()

    logger.info(
        "Cache metrics retrieved", extra={"admin_id": str(current_admin.user_id)}
    )
    return cache_stats


@router.post("/cache/clear")
async def clear_cache(
    current_admin: CurrentAdminUser,
    redis_session: RedisSession,
    pattern: str = "*",
) -> dict[str, Any]:
    """Clear cache entries matching a pattern.

    Requires admin authentication.

    Args:
        pattern: Redis pattern for keys to clear (default: "*" for all keys)

    Returns:
        Number of cache entries cleared
    """
    cache_service = CacheService(redis_session)
    cleared_count = await cache_service.clear_pattern(pattern)

    logger.info(
        "Cache cleared",
        extra={
            "admin_id": str(current_admin.user_id),
            "pattern": pattern,
            "cleared_count": cleared_count,
        },
    )

    return {
        "message": "Cache cleared successfully",
        "pattern": pattern,
        "cleared_count": cleared_count,
    }


@router.get("/system")
async def get_system_metrics(
    current_admin: CurrentAdminUser,
) -> dict[str, Any]:
    """Get system resource metrics.

    Requires admin authentication.

    Returns:
        System resource usage metrics
    """
    # Get system metrics
    cpu_percent = psutil.cpu_percent(interval=1)
    memory = psutil.virtual_memory()
    disk = psutil.disk_usage("/")

    # Get process-specific metrics
    process = psutil.Process(os.getpid())
    process_memory = process.memory_info()

    metrics: dict[str, Any] = {
        "timestamp": datetime.now(UTC).isoformat(),
        "system": {
            "cpu_usage_percent": cpu_percent,
            "memory_total_gb": round(memory.total / (1024**3), 2),
            "memory_used_gb": round(memory.used / (1024**3), 2),
            "memory_usage_percent": memory.percent,
            "disk_total_gb": round(disk.total / (1024**3), 2),
            "disk_used_gb": round(disk.used / (1024**3), 2),
            "disk_usage_percent": round((disk.used / disk.total) * 100, 2),
        },
        "process": {
            "memory_rss_mb": round(process_memory.rss / (1024**2), 2),
            "memory_vms_mb": round(process_memory.vms / (1024**2), 2),
            "cpu_percent": process.cpu_percent(),
            "num_threads": process.num_threads(),
            "open_files": len(process.open_files()),
        },
        "uptime_seconds": int(process.create_time()),
    }

    logger.info(
        "System metrics retrieved", extra={"admin_id": str(current_admin.user_id)}
    )
    return metrics


@router.get("/health/detailed")
async def get_detailed_health_metrics(
    current_admin: CurrentAdminUser,
    redis_session: RedisSession,
) -> dict[str, Any]:
    """Get detailed health metrics for all services.

    Requires admin authentication.

    Returns:
        Detailed health status for all service dependencies
    """
    health_status: dict[str, Any] = {
        "timestamp": datetime.now(UTC).isoformat(),
        "overall_status": "healthy",
        "services": {},
    }

    # Check Redis health
    try:
        start_time = asyncio.get_event_loop().time()
        await redis_session.ping()
        redis_response_time = (asyncio.get_event_loop().time() - start_time) * 1000

        cache_service = CacheService(redis_session)
        cache_stats = await cache_service.get_cache_stats()

        health_status["services"]["redis"] = {
            "status": "healthy",
            "response_time_ms": round(redis_response_time, 2),
            "memory_usage": cache_stats.get("memory_usage_human", "unknown"),
            "connected_clients": cache_stats.get("connected_clients", 0),
            "hit_rate_percent": cache_stats.get("hit_rate", 0.0),
        }

    except Exception as e:
        health_status["services"]["redis"] = {
            "status": "unhealthy",
            "error": str(e),
        }
        health_status["overall_status"] = "degraded"

    # Check database health (placeholder - would need actual DB connection)
    health_status["services"]["database"] = {
        "status": "healthy",
        "response_time_ms": 15.2,
        "active_connections": 8,
        "max_connections": 20,
    }

    # Add application metrics
    health_status["application"] = {
        "version": "1.0.0",
        "environment": "production",
        "features": {
            "authentication": "enabled",
            "caching": "enabled",
            "monitoring": "enabled",
            "security_headers": "enabled",
        },
    }

    logger.info(
        "Detailed health metrics retrieved",
        extra={
            "admin_id": str(current_admin.user_id),
            "status": health_status["overall_status"],
        },
    )

    return health_status
