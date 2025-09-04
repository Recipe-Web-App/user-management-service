"""OAuth2 token introspection response data schema."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class OAuth2IntrospectionData(BaseSchemaModel):
    """OAuth2 token introspection response data."""

    active: bool = Field(..., description="Whether the token is active")
    client_id: str | None = Field(None, description="Client identifier")
    username: str | None = Field(None, description="Username")
    sub: str | None = Field(None, description="Subject (user ID)")
    scope: str | None = Field(None, description="Space-delimited scopes")
    token_type: str | None = Field(None, description="Token type")
    exp: int | None = Field(None, description="Expiration timestamp")
    iat: int | None = Field(None, description="Issued at timestamp")
    aud: list[str] | None = Field(None, description="Audience")
    iss: str | None = Field(None, description="Issuer")

    @property
    def scopes(self) -> list[str]:
        """Get scopes as a list."""
        if not self.scope:
            return []
        return [scope.strip() for scope in self.scope.split() if scope.strip()]

    @property
    def user_id(self) -> str | None:
        """Get user ID from sub or username."""
        return self.sub or self.username
