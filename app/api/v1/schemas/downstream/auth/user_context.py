"""User authentication context schema."""

from datetime import datetime

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserContext(BaseSchemaModel):
    """User authentication context."""

    user_id: str = Field(..., description="User identifier")
    scopes: list[str] = Field(default_factory=list, description="User scopes")
    client_id: str | None = Field(None, description="OAuth2 client identifier")
    token_type: str = Field(default="Bearer", description="Token type")
    authenticated_at: datetime | None = Field(
        None, description="Authentication timestamp"
    )

    def has_scope(self, required_scope: str) -> bool:
        """Check if user has required scope."""
        return required_scope in self.scopes

    def has_any_scope(self, required_scopes: list[str]) -> bool:
        """Check if user has any of the required scopes."""
        return bool(set(required_scopes) & set(self.scopes))

    def has_all_scopes(self, required_scopes: list[str]) -> bool:
        """Check if user has all required scopes."""
        return set(required_scopes).issubset(set(self.scopes))
