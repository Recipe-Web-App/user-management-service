"""User login response schemas."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel
from app.api.v1.schemas.common.token import Token
from app.api.v1.schemas.common.user import User


class UserLoginResponse(BaseSchemaModel):
    """Response schema for successful user login."""

    user: User = Field(
        ...,
        description="The logged-in user's information",
    )
    token: Token = Field(
        ...,
        description="Access token information for the logged-in user",
    )
