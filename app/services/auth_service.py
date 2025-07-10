"""Authentication service for user management."""

from datetime import UTC, datetime, timedelta
from uuid import uuid4

from fastapi import HTTPException, status
from jose import jwt
from passlib.context import CryptContext

from app.api.v1.schemas.common.user import User as UserSchema
from app.api.v1.schemas.request.user_registration_request import UserRegistrationRequest
from app.api.v1.schemas.response.user_registration_response import (
    UserRegistrationResponse,
)
from app.core.config import settings
from app.core.logging import get_logger
from app.db.database_session import DatabaseSession
from app.db.models.user.user import User as UserModel
from app.enums.token_type import TokenType
from app.utils.security import SensitiveData

# Password hashing context
_pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")

_log = get_logger(__name__)


class AuthService:
    """Service for authentication operations."""

    def __init__(self, db: DatabaseSession) -> None:
        """Initialize auth service with database session."""
        self.db = db

    async def register_user(
        self, user_data: UserRegistrationRequest
    ) -> UserRegistrationResponse:
        """Register a new user.

        Args:
            user_data: User registration data

        Returns:
            UserRegistrationResponse: Registration result with user data and token

        Raises:
            HTTPException: If username or email already exists
        """
        existing_user = await self.db.get_user_by_username(user_data.username)
        if existing_user:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Username already registered",
            )

        existing_user = await self.db.get_user_by_email(user_data.email)
        if existing_user:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Email already registered",
            )

        user = await self._create_user(user_data)
        access_token = self._create_access_token(user)

        user_response = UserSchema.model_validate(user)

        return UserRegistrationResponse(
            user=user_response,
            access_token=SensitiveData(access_token),
            token_type=TokenType.BEARER,
        )

    def _get_password_hash(self, password: str) -> str:
        """Hash a password.

        Args:
            password: Plain text password

        Returns:
            str: Hashed password
        """
        return _pwd_context.hash(password)

    def _create_access_token(self, user: UserModel) -> str:
        """Create JWT access token for user.

        Args:
            user: User to create token for

        Returns:
            str: JWT token
        """
        data = {"sub": str(user.user_id), "username": user.username}
        to_encode = data.copy()
        expire = datetime.now(UTC) + timedelta(
            minutes=settings.access_token_expire_minutes
        )
        to_encode.update({"exp": expire})
        return jwt.encode(
            to_encode, settings.jwt_secret_key, algorithm=settings.jwt_signing_algorithm
        )

    async def _create_user(self, user_data: UserRegistrationRequest) -> UserModel:
        """Create a new user in the database.

        Args:
            user_data: User registration data

        Returns:
            User: Created user
        """
        # Extract password from SecretStr
        password = user_data.password.get_secret_value()
        hashed_password = self._get_password_hash(password)
        user = UserModel(
            user_id=uuid4(),
            username=user_data.username,
            email=user_data.email,
            password_hash=hashed_password,
            full_name=user_data.full_name,
            bio=user_data.bio,
            is_active=True,
        )

        return await self.db.create_user(user)
