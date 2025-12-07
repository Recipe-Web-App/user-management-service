"""OAuth2 access token data schema."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class OAuth2TokenData(BaseSchemaModel):
    """OAuth2 access token data."""

    access_token: str = Field(..., description="JWT access token")
    token_type: str = Field(default="Bearer", description="Token type")
    expires_in: int = Field(..., description="Token expiration time in seconds")
    refresh_token: str | None = Field(None, description="Refresh token")
    scope: str | None = Field(None, description="Space-delimited scopes")
    id_token: str | None = Field(None, description="OpenID Connect ID token")

    @property
    def scopes(self) -> list[str]:
        """Get scopes as a list."""
        if not self.scope:
            return []
        return [scope.strip() for scope in self.scope.split() if scope.strip()]
