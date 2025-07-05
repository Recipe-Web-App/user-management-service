"""Logging setup and configuration using Python's standard logging module.

This module configures structured logging for the entire application,
providing both console and file logging with request ID support and
JSON formatting for structured log analysis.
"""

import json
import logging
import logging.handlers
import sys
from pathlib import Path
from typing import ClassVar

from app.core.config import settings


class RequestIdFilter(logging.Filter):
    """Filter to add request_id to log records."""

    def filter(self, record: logging.LogRecord) -> bool:
        """Filter and modify log records to include request_id."""
        if not hasattr(record, "request_id"):
            record.request_id = "NULL"
        return True


class JsonFormatter(logging.Formatter):
    """Custom JSON formatter for structured logging."""

    def format(self, record: logging.LogRecord) -> str:
        """Format a log record as a JSON string."""
        log_entry = {
            "timestamp": self.formatTime(record),
            "level": record.levelname,
            "logger": record.name,
            "request_id": getattr(record, "request_id", "NULL"),
            "msg": record.getMessage(),
        }
        return json.dumps(log_entry)


class PrettyFormatter(logging.Formatter):
    """Pretty formatter for console output with colors and request ID handling."""

    # ANSI color codes
    COLORS: ClassVar[dict[str, str]] = {
        "DEBUG": "\033[36m",  # Cyan
        "INFO": "\033[32m",  # Green
        "WARNING": "\033[33m",  # Yellow
        "ERROR": "\033[31m",  # Red
        "CRITICAL": "\033[35m",  # Magenta
        "RESET": "\033[0m",  # Reset
        "GREEN": "\033[32m",  # Green for timestamp
        "BLUE": "\033[34m",  # Blue for request_id
        "CYAN": "\033[36m",  # Cyan for logger info
    }

    def format(self, record: logging.LogRecord) -> str:
        """Format a log record for console display with colors."""
        # Handle request_id formatting
        request_id = getattr(record, "request_id", "NULL")
        if request_id in ("NULL", "-"):
            colored_request_id = ""
            colored_separator = ""
        elif request_id:
            colored_request_id = (
                f"{self.COLORS['BLUE']}{request_id}{self.COLORS['RESET']}"
            )
            colored_separator = " | "
        else:
            colored_request_id = ""
            colored_separator = ""

        # Get color for log level
        level_color = self.COLORS.get(record.levelname, "")
        level_reset = self.COLORS["RESET"]

        # Format with colors
        return (
            f"{self.COLORS['GREEN']}{self.formatTime(record)}{self.COLORS['RESET']} | "
            f"{level_color}{record.levelname:<8}{level_reset} | "
            f"{self.COLORS['CYAN']}{record.name}:{record.funcName}:"
            f"{record.lineno}{self.COLORS['RESET']}"
            f"{colored_separator}{colored_request_id}"
            f" | {level_color}{record.getMessage()}{level_reset}"
        )


def configure_logging() -> None:
    """Configure global application logging using settings-based config."""
    # Disable noisy HTTP logging
    logging.getLogger("httpx").setLevel(logging.WARNING)
    logging.getLogger("httpcore").setLevel(logging.WARNING)
    logging.getLogger("httpcore.connection").setLevel(logging.WARNING)
    logging.getLogger("httpcore.http11").setLevel(logging.WARNING)

    # Create logs directory if it doesn't exist
    log_dir = Path("./logs")
    log_dir.mkdir(exist_ok=True)

    # Clear existing handlers
    root_logger = logging.getLogger()
    for handler in root_logger.handlers[:]:
        root_logger.removeHandler(handler)

    # Add request_id filter to root logger
    root_logger.addFilter(RequestIdFilter())

    # Configure handlers based on settings
    for sink in settings.logging_sinks:
        if sink.sink == "sys.stdout":
            # Console handler with pretty formatting
            console_handler = logging.StreamHandler(sys.stdout)
            console_handler.setLevel(getattr(logging, sink.level or "INFO"))
            console_handler.setFormatter(PrettyFormatter())
            root_logger.addHandler(console_handler)
        elif isinstance(sink.sink, str) and sink.sink.endswith(".log"):
            # File handler with JSON formatting
            file_handler = logging.handlers.RotatingFileHandler(
                sink.sink, maxBytes=10 * 1024 * 1024, backupCount=5  # 10MB
            )
            file_handler.setLevel(getattr(logging, sink.level or "DEBUG"))
            file_handler.setFormatter(JsonFormatter())
            root_logger.addHandler(file_handler)

    # Set root logger level to minimum of all handlers
    if root_logger.handlers:
        min_level = min(handler.level for handler in root_logger.handlers)
        root_logger.setLevel(min_level)


def get_logger(name: str | None = None) -> logging.Logger:
    """Retrieve a configured logger instance.

    Args:
        name: Optional logical name to bind to the logger.
              If None, uses the calling module's name.

    Returns:
        logging.Logger: A configured logger instance.

    """
    return logging.getLogger(name or __name__)
