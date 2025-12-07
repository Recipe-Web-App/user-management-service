"""User account deletion response schema definitions."""

from datetime import datetime
from uuid import UUID

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserAccountDeleteRequestResponse(BaseSchemaModel):
    """Response schema for account deletion request."""

    user_id: UUID = Field(..., description="User ID")
    confirmation_token: str = Field(..., description="Confirmation token for deletion")
    expires_at: datetime = Field(..., description="Token expiration time")
