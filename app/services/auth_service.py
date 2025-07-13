"""Authentication service for user management."""

from datetime import UTC, datetime, timedelta
from uuid import uuid4

from fastapi import HTTPException, status
from jose import jwt
from passlib.context import CryptContext

from app.api.v1.schemas.common.token import Token
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
from app.services.session_service import SessionData, SessionService

_log = get_logger(__name__)

# Password hashing context
_pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")


class AuthService:
    """Service for authentication operations."""

    def __init__(self, db: DatabaseSession) -> None:
        """Initialize auth service with database session."""
        self.db = db
        self.session_service = SessionService()

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
        _log.info(f"Registering user: {user_data}")
        existing_user = await self.db.get_user_by_username(user_data.username)
        if existing_user:
            _log.warning(f"Username already registered: {existing_user.username}")
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Username already registered",
            )

        existing_user = await self.db.get_user_by_email(user_data.email)
        if existing_user:
            _log.warning(f"Email already registered: {existing_user.email}")
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Email already registered",
            )

        user = await self._create_user(user_data)
        _log.info(f"User created: {user}")
        token = self._create_access_token(user)

        # Create session for the new user
        await self._create_user_session(user)

        user_response = UserSchema.model_validate(user)
        _log.debug(f"Transformed user model to response: {user_response}")

        return UserRegistrationResponse(
            user=user_response,
            token=token,
        )

    def _get_password_hash(self, password: str) -> str:
        """Hash a password.

        Args:
            password: Plain text password

        Returns:
            str: Hashed password
        """
        return _pwd_context.hash(password)

    def _verify_password(self, plain_password: str, hashed_password: str) -> bool:
        """Verify a password against its hash.

        Args:
            plain_password: Plain text password to verify
            hashed_password: Hashed password to check against

        Returns:
            bool: True if password matches, False otherwise
        """
        return _pwd_context.verify(plain_password, hashed_password)

    def _create_access_token(self, user: UserModel) -> Token:
        """Create JWT access token for user.

        Args:
            user: User to create token for

        Returns:
            Token: JWT token with expiration information
        """
        data = {"sub": str(user.user_id), "username": user.username}
        to_encode = data.copy()
        expire = datetime.now(UTC) + timedelta(
            minutes=settings.access_token_expire_minutes
        )
        to_encode.update({"exp": expire})

        access_token = jwt.encode(
            to_encode, settings.jwt_secret_key, algorithm=settings.jwt_signing_algorithm
        )

        # Calculate expiration time in seconds
        expires_in = settings.access_token_expire_minutes * 60

        return Token(
            access_token=access_token,
            token_type=TokenType.BEARER,
            expires_in=expires_in,
        )

    async def _create_user_session(self, user: UserModel) -> SessionData:
        """Create a session for a user.

        Args:
            user: User to create session for

        Returns:
            SessionData: Created session data
        """
        session_metadata = {
            "username": user.username,
            "email": user.email,
            "login_method": "registration",
        }

        session = await self.session_service.create_session(
            user_id=str(user.user_id),
            ttl_seconds=settings.access_token_expire_minutes * 60,
            metadata=session_metadata,
        )

        _log.info(f"Created session for user {user.username}: {session.session_id}")
        return session

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
