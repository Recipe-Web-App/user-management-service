"""User statistics response schema.

Provides the response model for user statistics endpoint.
"""

from pydantic import BaseModel, Field


class UserStatsResponse(BaseModel):
    """Response schema for user statistics."""

    total_users: int = Field(..., description="Total number of registered users.")
    active_users: int = Field(..., description="Number of active users.")
    recently_registered: int = Field(
        ..., description="Number of users registered in the last 30 days."
    )
    deactivated_users: int = Field(..., description="Number of deactivated users.")
    verified_users: int = Field(
        ..., description="Number of users with verified email addresses."
    )
    admin_users: int = Field(..., description="Number of admin users.")
    users_online: int = Field(..., description="Number of users currently online.")
    average_registration_rate: float = Field(
        ..., description="Average daily user registrations."
    )
    user_retention_rate: float = Field(
        ..., description="User retention rate as percentage."
    )
