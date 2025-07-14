"""Authentication service for user management."""

from datetime import UTC, datetime, timedelta
from uuid import uuid4

import redis.exceptions
from fastapi import HTTPException, status
from jose import JWTError, jwt
from passlib.context import CryptContext
from sqlalchemy.exc import DisconnectionError
from sqlalchemy.exc import TimeoutError as SQLTimeoutError

from app.api.v1.schemas.common.token import Token
from app.api.v1.schemas.common.user import User as UserSchema
from app.api.v1.schemas.request.user_login_request import UserLoginRequest
from app.api.v1.schemas.request.user_password_reset_confirm_request import (
    UserPasswordResetConfirmRequest,
)
from app.api.v1.schemas.request.user_password_reset_request import (
    UserPasswordResetRequest,
)
from app.api.v1.schemas.request.user_refresh_request import UserRefreshRequest
from app.api.v1.schemas.request.user_registration_request import UserRegistrationRequest
from app.api.v1.schemas.response.user_login_response import UserLoginResponse
from app.api.v1.schemas.response.user_logout_response import UserLogoutResponse
from app.api.v1.schemas.response.user_password_reset_confirm_response import (
    UserPasswordResetConfirmResponse,
)
from app.api.v1.schemas.response.user_password_reset_response import (
    UserPasswordResetResponse,
)
from app.api.v1.schemas.response.user_refresh_response import UserRefreshResponse
from app.api.v1.schemas.response.user_registration_response import (
    UserRegistrationResponse,
)
from app.core.config import settings
from app.core.logging import get_logger
from app.db.redis.models.session_data import SessionData
from app.db.redis.redis_database_session import RedisDatabaseSession
from app.db.sql.models.user.user import User as UserModel
from app.db.sql.sql_database_session import SqlDatabaseSession
from app.enums.token_type import TokenType
from app.exceptions.custom_exceptions.database_exceptions import DatabaseError
from app.utils.security import SensitiveData

_log = get_logger(__name__)

# Password hashing context
_pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")


class AuthService:
    """Service for authentication operations."""

    def __init__(
        self, db: SqlDatabaseSession, redis_session: RedisDatabaseSession
    ) -> None:
        """Initialize auth service with database session."""
        self.db = db
        self.redis_session = redis_session

    async def register_user(
        self, user_data: UserRegistrationRequest
    ) -> UserRegistrationResponse:
        """Register a new user.

        Args:
            user_data: User registration data

        Returns:
            UserRegistrationResponse: Registration result with user data and token

        Raises:
            HTTPException: If username or email already exists, or database services
            are unavailable
        """
        _log.info(f"Registering user: {user_data}")

        try:
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
        except DatabaseError as e:
            _log.error(f"Database error during user registration: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            _log.error(f"Database connection error during user registration: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Database service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e

        token = self._create_tokens(user)

        # Create session for the new user
        try:
            await self._create_user_session(user)
        except redis.ConnectionError as e:
            _log.error(f"Redis connection error during registration: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "User registered successfully, but session service is temporarily "
                    "unavailable. Please try logging in again later."
                ),
            ) from e

        user_response = UserSchema.model_validate(user)
        _log.debug(f"Transformed user model to response: {user_response}")

        return UserRegistrationResponse(
            user=user_response,
            token=token,
        )

    async def _authenticate_user(self, login_data: UserLoginRequest) -> UserModel:
        """Authenticate user credentials.

        Args:
            login_data: User login credentials

        Returns:
            UserModel: Authenticated user

        Raises:
            HTTPException: If authentication fails
        """
        # Try to find user by username or email
        user = None
        if login_data.username:
            user = await self.db.get_user_by_username(login_data.username)
        elif login_data.email:
            user = await self.db.get_user_by_email(login_data.email)

        if not user:
            login_identifier = login_data.username or login_data.email
            _log.warning(f"Login failed: User not found - {login_identifier}")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid credentials",
            )

        # Check if user is active
        if not user.is_active:
            _log.warning(f"Login failed: Inactive user - {user.username}")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Account is deactivated",
            )

        # Verify password
        password = login_data.password.get_secret_value()
        password_hash = user.password_hash.get_raw_value() if user.password_hash else ""
        if not self._verify_password(password, password_hash):
            _log.warning(f"Login failed: Invalid password for user - {user.username}")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid credentials",
            )

        return user

    async def _check_existing_session(self, user: UserModel) -> None:
        """Check if user already has an active session.

        Args:
            user: User to check

        Raises:
            HTTPException: If user already has an active session
        """
        user_sessions = await self.redis_session.get_user_sessions(str(user.user_id))
        if user_sessions:
            _log.warning(f"Login failed: User already logged in - {user.username}")
            raise HTTPException(
                status_code=status.HTTP_409_CONFLICT,
                detail=(
                    "User is already logged in. "
                    "Please log out before logging in again."
                ),
            )

    async def login_user(self, login_data: UserLoginRequest) -> UserLoginResponse:
        """Log in a user.

        Args:
            login_data: User login credentials

        Returns:
            UserLoginResponse: Login result with user data and access token

        Raises:
            HTTPException: If credentials are invalid, user is inactive, or database
            services are unavailable
        """
        _log.info(f"Login attempt for user: {login_data.username or login_data.email}")

        try:
            # Authenticate user
            user = await self._authenticate_user(login_data)

            # Check for existing session
            await self._check_existing_session(user)

            # Create access token
            token = self._create_tokens(user)

            # Create or update user session
            try:
                await self._create_user_session(user, login_method="login")
            except redis.ConnectionError as e:
                _log.error(f"Redis connection error during login: {e}")
                raise HTTPException(
                    status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                    detail=(
                        "Login successful, but session service is temporarily "
                        "unavailable. Please try again later."
                    ),
                ) from e

            user_response = UserSchema.model_validate(user)
            _log.debug(f"Transformed user model to response: {user_response}")

            return UserLoginResponse(
                user=user_response,
                token=token,
            )

        except HTTPException:
            # Re-raise HTTP exceptions as they are already properly formatted
            raise
        except DatabaseError as e:
            _log.error(f"Database error during user login: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            _log.error(f"Database connection error during user login: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Database service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e
        except Exception as e:
            _log.error(f"Unexpected error during user login: {e}")
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="An unexpected error occurred during login",
            ) from e

    async def refresh_token(
        self, refresh_data: UserRefreshRequest
    ) -> UserRefreshResponse:
        """Refresh access token using refresh token.

        Args:
            refresh_data: Refresh token data

        Returns:
            UserRefreshResponse: New access token

        Raises:
            HTTPException: If refresh token is invalid, expired, or user not found
        """
        _log.info("Token refresh attempt")

        try:
            # Decode and validate refresh token
            payload = self._decode_refresh_token(refresh_data.refresh_token)
            user_id = payload.get("sub")
            token_type = payload.get("type")

            if not user_id or token_type != "refresh":  # nosec
                _log.warning(
                    "Invalid refresh token: missing user ID or wrong token type"
                )
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Invalid refresh token",
                )

            # Get user from database
            user = await self.db.get_user_by_id(user_id)
            if not user:
                _log.warning(f"Refresh failed: User not found - {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Invalid refresh token",
                )

            # Check if user is active
            if not user.is_active:
                _log.warning(f"Refresh failed: Inactive user - {user.username}")
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Account is deactivated",
                )

            # Check if user has active session
            user_sessions = await self.redis_session.get_user_sessions(
                str(user.user_id)
            )
            if not user_sessions:
                _log.warning(
                    f"Refresh failed: No active session for user - {user.username}"
                )
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="No active session found",
                )

            # Create new access token (without refresh token for refresh response)
            access_token = self._create_access_token(user)
            _log.info(f"Token refreshed successfully for user: {user.username}")

            return UserRefreshResponse(
                message="Token refreshed successfully",
                token=access_token,
            )
        except JWTError as e:
            _log.warning(f"JWT decode error during token refresh: {e}")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid refresh token",
            ) from e
        except DatabaseError as e:
            _log.error(f"Database error during token refresh: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            _log.error(f"Database connection error during token refresh: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Database service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e
        except redis.ConnectionError as e:
            _log.error(f"Redis connection error during token refresh: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Session service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e

    async def request_password_reset(
        self, reset_data: UserPasswordResetRequest
    ) -> UserPasswordResetResponse:
        """Request password reset for a user.

        Args:
            reset_data: Password reset request data

        Returns:
            UserPasswordResetResponse: Password reset request result

        Raises:
            HTTPException: If user not found or service unavailable
        """
        _log.info(f"Password reset request for email: {reset_data.email}")

        try:
            # Find user by email
            user = await self.db.get_user_by_email(reset_data.email)
            if not user:
                _log.warning(
                    f"Password reset failed: User not found - {reset_data.email}"
                )
                # Don't reveal if user exists or not for security
                return UserPasswordResetResponse(
                    message="If the email exists, a password reset link has been sent",
                    email_sent=True,
                )

            # Check if user is active
            if not user.is_active:
                _log.warning(f"Password reset failed: Inactive user - {user.username}")
                return UserPasswordResetResponse(
                    message="If the email exists, a password reset link has been sent",
                    email_sent=True,
                )

            # Create password reset token
            reset_token = self._create_password_reset_token(user)
            _log.info(f"Password reset token created for user: {user.username}")

            # TODO: Send email with reset token
            # For now, we'll just log the token (in production, send via email)
            _log.info(f"Password reset token for {user.email}: {reset_token}")

            return UserPasswordResetResponse(
                message="If the email exists, a password reset link has been sent",
                email_sent=True,
            )

        except DatabaseError as e:
            _log.error(f"Database error during password reset request: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            _log.error(f"Database connection error during password reset request: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Database service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e
        except Exception as e:
            _log.error(f"Unexpected error during password reset request: {e}")
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="An unexpected error occurred during password reset request",
            ) from e

    async def confirm_password_reset(
        self, confirm_data: UserPasswordResetConfirmRequest
    ) -> UserPasswordResetConfirmResponse:
        """Confirm password reset using reset token.

        Args:
            confirm_data: Password reset confirmation data

        Returns:
            UserPasswordResetConfirmResponse: Password reset confirmation result

        Raises:
            HTTPException: If reset token is invalid, expired, or user not found
        """
        _log.info("Password reset confirmation attempt")

        try:
            # Decode and validate reset token
            payload = self._decode_password_reset_token(confirm_data.reset_token)
            user_id = payload.get("sub")
            token_type = payload.get("type")

            if not user_id or token_type != "password_reset":  # nosec
                _log.warning(
                    "Invalid password reset token: missing user ID or wrong token type"
                )
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Invalid or expired reset token",
                )

            # Get user from database
            user = await self.db.get_user_by_id(user_id)
            if not user:
                _log.warning(f"Password reset failed: User not found - {user_id}")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Invalid or expired reset token",
                )

            # Check if user is active
            if not user.is_active:
                _log.warning(f"Password reset failed: Inactive user - {user.username}")
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="Account is deactivated",
                )

            # Update user password
            new_password = confirm_data.new_password.get_secret_value()
            hashed_password = self._get_password_hash(new_password)

            # Update user in database
            if user.password_hash is None:
                user.password_hash = SensitiveData(hashed_password)
            else:
                user.password_hash.set_raw_value(hashed_password)
            await self.db.commit()
            await self.db.refresh(user)

            # Invalidate all user sessions (force re-login with new password)
            try:
                await self.redis_session.invalidate_user_sessions(str(user.user_id))
                _log.info(
                    "Invalidated sessions for user after password reset: "
                    f"{user.username}"
                )
            except redis.ConnectionError as e:
                _log.warning(f"Could not invalidate sessions after password reset: {e}")

            _log.info(f"Password reset successfully for user: {user.username}")

            return UserPasswordResetConfirmResponse(
                message="Password reset successfully",
                password_updated=True,
            )

        except JWTError as e:
            _log.warning(f"JWT decode error during password reset: {e}")
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Invalid or expired reset token",
            ) from e
        except DatabaseError as e:
            _log.error(f"Database error during password reset: {e}")
            raise HTTPException(
                status_code=e.status_code,
                detail=str(e),
            ) from e
        except (DisconnectionError, SQLTimeoutError) as e:
            _log.error(f"Database connection error during password reset: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Database service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e

    async def logout_user(self, user_id: str) -> UserLogoutResponse:
        """Log out a user by invalidating their sessions.

        Args:
            user_id: The user ID to logout

        Returns:
            UserLogoutResponse: Logout result with confirmation

        Raises:
            HTTPException: If logout fails due to service unavailability
        """
        _log.info(f"Logout attempt for user: {user_id}")

        try:
            # Invalidate all sessions for the user
            sessions_invalidated = await self.redis_session.invalidate_user_sessions(
                user_id
            )

            if sessions_invalidated > 0:
                _log.info(
                    f"User logged out successfully: {user_id} "
                    f"(invalidated {sessions_invalidated} sessions)"
                )
            else:
                _log.info(f"User logged out (no active sessions found): {user_id}")

            return UserLogoutResponse(
                message="User logged out successfully",
                session_invalidated=sessions_invalidated > 0,
            )
        except redis.ConnectionError as e:
            _log.error(f"Redis connection error during logout: {e}")
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=(
                    "Logout failed: Session service is temporarily unavailable. "
                    "Please try again later."
                ),
            ) from e
        except Exception as e:
            _log.error(f"Unexpected error during logout: {e}")
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="An unexpected error occurred during logout",
            ) from e

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
            refresh_token=None,
            token_type=TokenType.BEARER,
            expires_in=expires_in,
        )

    def _create_refresh_token(self, user: UserModel) -> str:
        """Create JWT refresh token for user.

        Args:
            user: User to create refresh token for

        Returns:
            str: JWT refresh token
        """
        data = {
            "sub": str(user.user_id),
            "username": user.username,
            "type": "refresh",  # nosec
        }
        to_encode = data.copy()
        expire = datetime.now(UTC) + timedelta(days=settings.refresh_token_expire_days)
        to_encode.update({"exp": expire})

        return jwt.encode(
            to_encode, settings.jwt_secret_key, algorithm=settings.jwt_signing_algorithm
        )

    def _create_password_reset_token(self, user: UserModel) -> str:
        """Create JWT password reset token for user.

        Args:
            user: User to create password reset token for

        Returns:
            str: JWT password reset token
        """
        data = {
            "sub": str(user.user_id),
            "username": user.username,
            "type": "password_reset",  # nosec
        }
        to_encode = data.copy()
        expire = datetime.now(UTC) + timedelta(
            minutes=settings.password_reset_token_expire_minutes
        )
        to_encode.update({"exp": expire})

        return jwt.encode(
            to_encode, settings.jwt_secret_key, algorithm=settings.jwt_signing_algorithm
        )

    def _create_tokens(self, user: UserModel) -> Token:
        """Create both access and refresh tokens for user.

        Args:
            user: User to create tokens for

        Returns:
            Token: JWT tokens with expiration information
        """
        access_token = self._create_access_token(user)
        refresh_token = self._create_refresh_token(user)

        return Token(
            access_token=access_token.access_token,
            refresh_token=refresh_token,
            token_type=TokenType.BEARER,
            expires_in=access_token.expires_in,
        )

    def _decode_refresh_token(self, refresh_token: str) -> dict:
        """Decode and validate refresh token.

        Args:
            refresh_token: JWT refresh token to decode

        Returns:
            dict: Decoded token payload

        Raises:
            jwt.JWTError: If token is invalid or expired
        """
        return jwt.decode(
            refresh_token,
            settings.jwt_secret_key,
            algorithms=[settings.jwt_signing_algorithm],
        )

    def _decode_password_reset_token(self, reset_token: str) -> dict:
        """Decode and validate password reset token.

        Args:
            reset_token: JWT password reset token to decode

        Returns:
            dict: Decoded token payload

        Raises:
            jwt.JWTError: If token is invalid or expired
        """
        return jwt.decode(
            reset_token,
            settings.jwt_secret_key,
            algorithms=[settings.jwt_signing_algorithm],
        )

    async def _create_user_session(
        self, user: UserModel, login_method: str = "registration"
    ) -> SessionData:
        """Create a session for a user.

        Args:
            user: User to create session for
            login_method: Method used to create the session (registration/login)

        Returns:
            SessionData: Created session data
        """
        session_metadata = {
            "username": user.username,
            "email": user.email,
            "login_method": login_method,
        }

        session = await self.redis_session.create_session(
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
