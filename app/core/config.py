"""Application configuration settings.

Defines and loads configuration variables and settings used across the application,
including environment-specific and default configurations.
"""

import json
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

    # JWT Authentication Settings
    JWT_SECRET_KEY: str = Field(..., alias="JWT_SECRET_KEY")
    JWT_SIGNING_ALGORITHM: str = Field(..., alias="JWT_SIGNING_ALGORITHM")
    ACCESS_TOKEN_EXPIRE_MINUTES: int = Field(..., alias="ACCESS_TOKEN_EXPIRE_MINUTES")

    # CORS Configuration
    ALLOWED_ORIGIN_HOSTS: str = Field(..., alias="ALLOWED_ORIGIN_HOSTS")
    ALLOWED_CREDENTIALS: bool = Field(..., alias="ALLOWED_CREDENTIALS")

    LOGGING_CONFIG_PATH: str = Field(
        str(
            (Path(__file__).parent.parent.parent / "config" / "logging.json").resolve(),
        ),
        alias="LOGGING_CONFIG_PATH",
    )

    _LOGGING_SINKS: list["LoggingSink"] = PrivateAttr(default_factory=list)

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        validate_default=True,
    )

    def __init__(self) -> None:
        """Load logging config after Pydantic initialization."""
        super().__init__()

        # Load logging configuration
        config_path = Path(self.LOGGING_CONFIG_PATH).expanduser().resolve()
        with config_path.open("r", encoding="utf-8") as f:
            config = json.load(f)
        sinks = config.get("sinks", [])
        self._LOGGING_SINKS = [
            LoggingSink.from_dict(s) for s in sinks if isinstance(s, dict)
        ]

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
    def allowed_origin_hosts(self) -> str:
        """Get allowed origin hosts for CORS."""
        return self.ALLOWED_ORIGIN_HOSTS

    @property
    def allowed_credentials(self) -> bool:
        """Get allowed credentials setting for CORS."""
        return self.ALLOWED_CREDENTIALS

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


settings = _Settings()
