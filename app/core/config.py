"""Application configuration settings.

Defines and loads configuration variables and settings used across the application,
including environment-specific and default configurations.
"""

import json
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

from pydantic import Field, PrivateAttr
from pydantic_settings import BaseSettings, SettingsConfigDict


class LoggingSink:
    """Represents a single logging sink configuration.

    Attributes:
        sink: The sink target (e.g., file path or sys.stdout).
        level: The log level for this sink (e.g., "INFO", "DEBUG").
        serialize: Whether to serialize logs as JSON.
        rotation: Log rotation policy (e.g., "10 MB").
        retention: Log retention policy (e.g., "10 days").
        compression: Compression for rotated logs (e.g., "zip").
        colorize: Enable colored output for console sinks.
        catch: Catch sink exceptions.
    """

    def __init__(  # noqa: PLR0913
        self,
        sink: Any,
        level: str | None = None,
        serialize: bool | None = None,
        rotation: str | None = None,
        retention: str | None = None,
        compression: str | None = None,
        colorize: bool | None = None,
        catch: bool | None = None,
    ) -> None:
        """Initialize LoggingSink."""
        self.sink = sink
        self.level = level
        self.serialize = serialize
        self.rotation = rotation
        self.retention = retention
        self.compression = compression
        self.colorize = colorize
        self.catch = catch

    @staticmethod
    def from_dict(data: dict[str, Any]) -> "LoggingSink":
        """Create a LoggingSink instance from a dictionary."""
        return LoggingSink(
            sink=data.get("sink"),
            level=data.get("level"),
            serialize=data.get("serialize"),
            rotation=data.get("rotation"),
            retention=data.get("retention"),
            compression=data.get("compression"),
            colorize=data.get("colorize"),
            catch=data.get("catch"),
        )


class _Settings(BaseSettings):
    """Application settings loaded from environment variables or .env file."""

    # Database Configuration
    POSTGRES_HOST: str = Field(..., alias="POSTGRES_HOST")
    POSTGRES_PORT: int = Field(..., alias="POSTGRES_PORT")
    POSTGRES_DB: str = Field(..., alias="POSTGRES_DB")
    POSTGRES_SCHEMA: str = Field(..., alias="POSTGRES_SCHEMA")

    # User Management Database User
    USER_MANAGEMENT_DB_USER: str = Field(..., alias="USER_MANAGEMENT_DB_USER")
    USER_MANAGEMENT_DB_PASSWORD: str = Field(..., alias="USER_MANAGEMENT_DB_PASSWORD")

    # JWT Authentication Settings (for backward compatibility)
    JWT_SECRET_KEY: str = Field(..., alias="JWT_SECRET_KEY")
    JWT_SIGNING_ALGORITHM: str = Field(..., alias="JWT_SIGNING_ALGORITHM")
    ACCESS_TOKEN_EXPIRE_MINUTES: int = Field(..., alias="ACCESS_TOKEN_EXPIRE_MINUTES")
    REFRESH_TOKEN_EXPIRE_DAYS: int = Field(..., alias="REFRESH_TOKEN_EXPIRE_DAYS")
    PASSWORD_RESET_TOKEN_EXPIRE_MINUTES: int = Field(
        ..., alias="PASSWORD_RESET_TOKEN_EXPIRE_MINUTES"
    )

    # OAuth2 Integration Settings
    JWT_SECRET: str = Field("", alias="JWT_SECRET")  # Shared secret with OAuth2 service
    OAUTH2_SERVICE_ENABLED: bool = Field(True, alias="OAUTH2_SERVICE_ENABLED")
    OAUTH2_SERVICE_TO_SERVICE_ENABLED: bool = Field(
        True, alias="OAUTH2_SERVICE_TO_SERVICE_ENABLED"
    )
    OAUTH2_INTROSPECTION_ENABLED: bool = Field(
        False, alias="OAUTH2_INTROSPECTION_ENABLED"
    )
    OAUTH2_CLIENT_ID: str = Field("", alias="OAUTH2_CLIENT_ID")
    OAUTH2_CLIENT_SECRET: str = Field("", alias="OAUTH2_CLIENT_SECRET")

    # CORS Configuration
    ALLOWED_ORIGIN_HOSTS: str = Field(..., alias="ALLOWED_ORIGIN_HOSTS")
    ALLOWED_CREDENTIALS: bool = Field(..., alias="ALLOWED_CREDENTIALS")

    # Redis Configuration
    REDIS_HOST: str = Field(..., alias="REDIS_HOST")
    REDIS_PORT: int = Field(..., alias="REDIS_PORT")
    REDIS_PASSWORD: str = Field(..., alias="REDIS_PASSWORD")
    REDIS_DB: int = Field(..., alias="REDIS_DB")

    # State Configuration
    _STARTUP_TIME: datetime = PrivateAttr(default_factory=lambda: datetime.now(UTC))

    LOGGING_CONFIG_PATH: str = Field(
        str(
            (Path(__file__).parent.parent.parent / "config" / "logging.json").resolve(),
        ),
        alias="LOGGING_CONFIG_PATH",
    )

    _LOGGING_SINKS: list["LoggingSink"] = PrivateAttr(default_factory=list)
    _OAUTH2_CONFIG: dict[str, Any] = PrivateAttr(default_factory=dict)

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        validate_default=True,
    )

    def __init__(self) -> None:
        """Load logging and OAuth2 config after Pydantic initialization."""
        super().__init__()

        # Load logging configuration
        config_path = Path(self.LOGGING_CONFIG_PATH).expanduser().resolve()
        with config_path.open("r", encoding="utf-8") as f:
            config = json.load(f)
        sinks = config.get("sinks", [])
        self._LOGGING_SINKS = [
            LoggingSink.from_dict(s) for s in sinks if isinstance(s, dict)
        ]

        # Load OAuth2 configuration
        oauth2_config_path = (
            Path(__file__).parent.parent.parent / "config" / "oauth2.json"
        )
        if oauth2_config_path.exists():
            with oauth2_config_path.open("r", encoding="utf-8") as f:
                self._OAUTH2_CONFIG = json.load(f)
        else:
            self._OAUTH2_CONFIG = {}

    @property
    def postgres_host(self) -> str:
        """Get PostgreSQL host."""
        return self.POSTGRES_HOST

    @property
    def postgres_port(self) -> int:
        """Get PostgreSQL port."""
        return self.POSTGRES_PORT

    @property
    def postgres_db(self) -> str:
        """Get PostgreSQL database name."""
        return self.POSTGRES_DB

    @property
    def postgres_schema(self) -> str:
        """Get PostgreSQL schema name."""
        return self.POSTGRES_SCHEMA

    @property
    def user_management_db_user(self) -> str:
        """Get user management database username."""
        return self.USER_MANAGEMENT_DB_USER

    @property
    def user_management_db_password(self) -> str:
        """Get user management database password."""
        return self.USER_MANAGEMENT_DB_PASSWORD

    @property
    def jwt_secret_key(self) -> str:
        """Get JWT secret key."""
        return self.JWT_SECRET_KEY

    @property
    def jwt_signing_algorithm(self) -> str:
        """Get JWT signing algorithm."""
        return self.JWT_SIGNING_ALGORITHM

    @property
    def access_token_expire_minutes(self) -> int:
        """Get access token expiration time in minutes."""
        return self.ACCESS_TOKEN_EXPIRE_MINUTES

    @property
    def refresh_token_expire_days(self) -> int:
        """Get refresh token expiration time in days."""
        return self.REFRESH_TOKEN_EXPIRE_DAYS

    @property
    def password_reset_token_expire_minutes(self) -> int:
        """Get password reset token expiration time in minutes."""
        return self.PASSWORD_RESET_TOKEN_EXPIRE_MINUTES

    @property
    def allowed_origin_hosts(self) -> str:
        """Get allowed origin hosts for CORS."""
        return self.ALLOWED_ORIGIN_HOSTS

    @property
    def allowed_credentials(self) -> bool:
        """Get allowed credentials setting for CORS."""
        return self.ALLOWED_CREDENTIALS

    @property
    def redis_host(self) -> str:
        """Get Redis host."""
        return self.REDIS_HOST

    @property
    def redis_port(self) -> int:
        """Get Redis port."""
        return self.REDIS_PORT

    @property
    def redis_password(self) -> str:
        """Get Redis password."""
        return self.REDIS_PASSWORD

    @property
    def redis_db(self) -> int:
        """Get Redis database number."""
        return self.REDIS_DB

    @property
    def logging_sinks(self) -> list["LoggingSink"]:
        """Get all configured logging sinks."""
        return self._LOGGING_SINKS

    @property
    def logging_stdout_sink(self) -> LoggingSink | None:
        """Get the stdout logging sink configuration."""
        return next(
            (sink for sink in self._LOGGING_SINKS if sink.sink == "sys.stdout"),
            None,
        )

    @property
    def logging_file_sink(self) -> LoggingSink | None:
        """Get the file logging sink configuration."""
        return next(
            (
                sink
                for sink in self._LOGGING_SINKS
                if isinstance(sink.sink, str) and sink.sink.endswith(".log")
            ),
            None,
        )

    @property
    def startup_time(self) -> datetime:
        """Get the startup time."""
        return self._STARTUP_TIME

    # OAuth2 Integration Properties
    @property
    def jwt_secret(self) -> str:
        """Get JWT secret for OAuth2 integration."""
        return self.JWT_SECRET

    @property
    def oauth2_service_enabled(self) -> bool:
        """Check if OAuth2 service integration is enabled."""
        return self.OAUTH2_SERVICE_ENABLED

    @property
    def oauth2_service_to_service_enabled(self) -> bool:
        """Check if OAuth2 service-to-service authentication is enabled."""
        return self.OAUTH2_SERVICE_TO_SERVICE_ENABLED

    @property
    def oauth2_introspection_enabled(self) -> bool:
        """Check if OAuth2 token introspection is enabled."""
        return self.OAUTH2_INTROSPECTION_ENABLED

    @property
    def oauth2_client_id(self) -> str:
        """Get OAuth2 client ID."""
        return self.OAUTH2_CLIENT_ID

    @property
    def oauth2_client_secret(self) -> str:
        """Get OAuth2 client secret."""
        return self.OAUTH2_CLIENT_SECRET

    @property
    def oauth2_authorization_url(self) -> str:
        """Get OAuth2 authorization URL."""
        return self._OAUTH2_CONFIG.get("oauth2_service_urls", {}).get(
            "authorization_url", ""
        )

    @property
    def oauth2_token_url(self) -> str:
        """Get OAuth2 token URL."""
        return self._OAUTH2_CONFIG.get("oauth2_service_urls", {}).get("token_url", "")

    @property
    def oauth2_introspection_url(self) -> str:
        """Get OAuth2 token introspection URL."""
        return self._OAUTH2_CONFIG.get("oauth2_service_urls", {}).get(
            "introspection_url", ""
        )

    @property
    def oauth2_userinfo_url(self) -> str:
        """Get OAuth2 userinfo URL."""
        return self._OAUTH2_CONFIG.get("oauth2_service_urls", {}).get(
            "userinfo_url", ""
        )

    @property
    def oauth2_default_scopes(self) -> list[str]:
        """Get OAuth2 default scopes as list."""
        return self._OAUTH2_CONFIG.get("scopes", {}).get(
            "default_scopes", ["openid", "profile"]
        )

    @property
    def oauth2_admin_scopes(self) -> list[str]:
        """Get OAuth2 admin scopes as list."""
        return self._OAUTH2_CONFIG.get("scopes", {}).get(
            "admin_scopes", ["openid", "profile", "admin"]
        )

    @property
    def oauth2_user_management_scopes(self) -> list[str]:
        """Get OAuth2 user management scopes as list."""
        return self._OAUTH2_CONFIG.get("scopes", {}).get(
            "user_management_scopes", ["openid", "profile", "user:read", "user:write"]
        )

    def get_effective_jwt_secret(self) -> str:
        """Get the effective JWT secret based on OAuth2 configuration."""
        if (
            self.oauth2_service_enabled
            and not self.oauth2_introspection_enabled
            and self.jwt_secret
        ):
            return self.jwt_secret
        return self.jwt_secret_key


settings = _Settings()
