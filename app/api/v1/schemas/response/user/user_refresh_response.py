"""User refresh token response schemas."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel
from app.api.v1.schemas.common.token import Token


class UserRefreshResponse(BaseSchemaModel):
    """Response schema for token refresh."""

    message: str = Field(
        ...,
        description="Success message",
        examples=["Token refreshed successfully"],
    )
    token: Token = Field(
        ...,
        description="New access token with expiration information",
    )
