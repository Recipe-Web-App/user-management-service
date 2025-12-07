"""Pagination dependency providers."""

from typing import Annotated

from fastapi import Depends, Query


def get_pagination_params(
    offset: Annotated[int, Query(ge=0, description="Number of items to skip")] = 0,
    limit: Annotated[
        int, Query(ge=1, le=100, description="Number of items to return")
    ] = 20,
) -> dict[str, int]:
    """Get pagination parameters."""
    return {"offset": offset, "limit": limit}


def get_search_params(
    query: Annotated[
        str, Query(min_length=1, max_length=100, description="Search query")
    ] = "",
    offset: Annotated[int, Query(ge=0, description="Number of items to skip")] = 0,
    limit: Annotated[
        int, Query(ge=1, le=100, description="Number of items to return")
    ] = 20,
) -> dict[str, int | str]:
    """Get search and pagination parameters."""
    return {"query": query, "offset": offset, "limit": limit}


# Type aliases for dependency injection
PaginationParams = Annotated[dict[str, int], Depends(get_pagination_params)]
SearchParams = Annotated[dict[str, int | str], Depends(get_search_params)]
