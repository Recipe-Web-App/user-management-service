"""OAuth2 client credentials configuration schema."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class OAuth2ClientCredentials(BaseSchemaModel):
    """OAuth2 client credentials configuration."""

    client_id: str = Field(..., description="OAuth2 client ID")
    client_secret: str = Field(..., description="OAuth2 client secret")
    default_scopes: list[str] = Field(
        default_factory=list, description="Default scopes"
    )
