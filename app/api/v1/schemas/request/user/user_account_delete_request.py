"""User account deletion request schema definitions."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel


class UserAccountDeleteRequest(BaseSchemaModel):
    """Request schema for confirming user account deletion."""

    confirmation_token: str = Field(
        ...,
        min_length=1,
        description="Confirmation token received from delete request",
    )
