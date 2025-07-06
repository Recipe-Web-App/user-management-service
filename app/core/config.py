"""Configuration management for the User Management Service.

This module provides configuration classes and settings management, including logging
sink configuration loaded from JSON files.
"""

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any


@dataclass(frozen=True)
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

    sink: Any
    level: str | None = None
    serialize: bool | None = None
    rotation: str | None = None
    retention: str | None = None
    compression: str | None = None
    colorize: bool | None = None
    catch: bool | None = None

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


class Settings:
    """Application settings, including logging sinks configuration.

    Manages application configuration settings, particularly logging configuration
    loaded from JSON files.
    """

    LOGGING_CONFIG_PATH: str = str(
        (Path(__file__).parent.parent.parent / "config" / "logging.json").resolve()
    )

    def __init__(self) -> None:
        """Initialize settings and load logging configuration."""
        config_path = Path(self.LOGGING_CONFIG_PATH).expanduser().resolve()
        with config_path.open("r", encoding="utf-8") as f:
            config: dict[str, Any] = json.load(f)
        sinks: list[dict[str, Any]] = config.get("sinks", [])
        self._logging_sinks = [LoggingSink.from_dict(s) for s in sinks]

    @property
    def logging_sinks(self) -> list[LoggingSink]:
        """Get all configured logging sinks."""
        return self._logging_sinks

    @property
    def logging_stdout_sink(self) -> LoggingSink | None:
        """Get the stdout logging sink configuration."""
        return next((s for s in self._logging_sinks if s.sink == "sys.stdout"), None)

    @property
    def logging_file_sink(self) -> LoggingSink | None:
        """Get the file logging sink configuration."""
        return next(
            (
                s
                for s in self._logging_sinks
                if isinstance(s.sink, str) and s.sink.endswith(".log")
            ),
            None,
        )


# Global settings instance for easy access throughout the application
settings = Settings()
