"""Social response schemas package."""

from .follow_response import FollowResponse
from .get_followed_users_response import GetFollowedUsersResponse

__all__ = [
    "FollowResponse",
    "GetFollowedUsersResponse",
]
