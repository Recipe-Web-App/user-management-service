"""User authentication context schema."""

from datetime import datetime

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserContext(BaseSchemaModel):
    """User or service authentication context.

    For user tokens: user_id is set, client_id may be set
    For service tokens: user_id is None, client_id is set
    """

    user_id: str | None = Field(
        None, description="User identifier (None for service tokens)"
    )
    scopes: list[str] = Field(default_factory=list, description="Token scopes")
    client_id: str | None = Field(None, description="OAuth2 client identifier")
    token_type: str = Field(default="Bearer", description="Token type")
    authenticated_at: datetime | None = Field(
        None, description="Authentication timestamp"
    )

    @property
    def is_service_token(self) -> bool:
        """Check if this context is from a service token (no user_id)."""
        return self.user_id is None and self.client_id is not None

    def has_scope(self, required_scope: str) -> bool:
        """Check if user has required scope."""
        return required_scope in self.scopes

    def has_any_scope(self, required_scopes: list[str]) -> bool:
        """Check if user has any of the required scopes."""
        return bool(set(required_scopes) & set(self.scopes))

    def has_all_scopes(self, required_scopes: list[str]) -> bool:
        """Check if user has all required scopes."""
        return set(required_scopes).issubset(set(self.scopes))
