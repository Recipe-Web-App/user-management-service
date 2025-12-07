"""User confirm account deletion response schema definitions."""

from datetime import datetime
from uuid import UUID

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserConfirmAccountDeleteResponse(BaseSchemaModel):
    """Response schema for confirming account deletion."""

    user_id: UUID = Field(..., description="User ID")
    deactivated_at: datetime = Field(..., description="Account deactivation time")
