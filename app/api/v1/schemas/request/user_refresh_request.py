"""User refresh token request schemas."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserRefreshRequest(BaseSchemaModel):
    """Request schema for token refresh."""

    refresh_token: str = Field(
        ...,
        description="Refresh token to exchange for new access token",
        examples=["eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."],
    )
