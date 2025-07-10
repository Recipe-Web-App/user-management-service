"""User registration response schemas."""

from pydantic import Field

from app.api.v1.schemas.base_schema_model import BaseSchemaModel
from app.api.v1.schemas.common.user import User
from app.enums.token_type import TokenType
from app.utils.security import SensitiveData


class UserRegistrationResponse(BaseSchemaModel):
    """Response schema for successful user registration."""

    user: User = Field(
        ...,
        description="The registered user's information",
    )
    access_token: SensitiveData = Field(
        ...,
        description="JWT access token for the registered user",
        examples=["eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."],
    )
    token_type: TokenType = Field(
        default=TokenType.BEARER,
        description="Type of authentication token",
        examples=["bearer"],
    )
