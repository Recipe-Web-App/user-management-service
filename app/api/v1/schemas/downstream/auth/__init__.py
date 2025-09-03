"""Downstream authentication schemas package."""

from app.api.v1.schemas.downstream.auth.jwt_token_payload import JWTTokenPayload
from app.api.v1.schemas.downstream.auth.oauth2_client_credentials import (
    OAuth2ClientCredentials,
)
from app.api.v1.schemas.downstream.auth.oauth2_introspection_data import (
    OAuth2IntrospectionData,
)
from app.api.v1.schemas.downstream.auth.oauth2_token_data import OAuth2TokenData
from app.api.v1.schemas.downstream.auth.user_context import UserContext

__all__ = [
    "JWTTokenPayload",
    "OAuth2ClientCredentials",
    "OAuth2IntrospectionData",
    "OAuth2TokenData",
    "UserContext",
]
