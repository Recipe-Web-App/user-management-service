"""JWT token payload data schema."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class JWTTokenPayload(BaseSchemaModel):
    """JWT token payload data."""

    iss: str | None = Field(None, description="Issuer")
    aud: list[str] | str | None = Field(None, description="Audience")
    sub: str | None = Field(None, description="Subject (user ID)")
    client_id: str | None = Field(None, description="Client identifier")
    user_id: str | None = Field(None, description="User identifier")
    scopes: list[str] | None = Field(None, description="Token scopes")
    type: str | None = Field(None, description="Token type")
    exp: int | None = Field(None, description="Expiration timestamp")
    iat: int | None = Field(None, description="Issued at timestamp")
    nbf: int | None = Field(None, description="Not before timestamp")
    jti: str | None = Field(None, description="JWT ID")

    @property
    def effective_user_id(self) -> str | None:
        """Get effective user ID from sub or user_id."""
        return self.sub or self.user_id

    @property
    def effective_scopes(self) -> list[str]:
        """Get effective scopes list."""
        return self.scopes or []
