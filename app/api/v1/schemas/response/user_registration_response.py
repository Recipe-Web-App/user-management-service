"""User registration response schemas."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel
from app.api.v1.schemas.common.token import Token
from app.api.v1.schemas.common.user import User


class UserRegistrationResponse(BaseSchemaModel):
    """Response schema for successful user registration."""

    user: User = Field(
        ...,
        description="The registered user's information",
    )
    token: Token | None = Field(
        None,
        description="Access token information for the registered user",
    )
