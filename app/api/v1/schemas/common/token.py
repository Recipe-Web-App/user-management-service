"""Token schema."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel
from app.enums.token_type import TokenType


class Token(BaseSchemaModel):
    """Token schema."""

    access_token: str = Field(
        ...,
        description="JWT access token for the registered user",
    )
    refresh_token: str | None = Field(
        None,
        description="JWT refresh token for obtaining new access tokens",
    )
    token_type: TokenType = Field(
        default=TokenType.BEARER,
        description="Type of authentication token",
        examples=["bearer"],
    )
    expires_in: int = Field(
        ...,
        description="Expiration time of the token in seconds",
    )
