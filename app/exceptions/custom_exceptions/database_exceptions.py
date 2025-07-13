"""Database-related custom exceptions."""

from http import HTTPStatus
from typing import ClassVar

import asyncpg


class DatabaseError(Exception):
    """Custom database error with HTTP status code."""

    # PostgreSQL error code constants using asyncpg error classes

    # Connection and availability issues (503)
    CONNECTION_EXCEPTION: ClassVar[str] = asyncpg.ConnectionFailureError.sqlstate
    SQLCLIENT_UNABLE_TO_ESTABLISH_SQLCONNECTION: ClassVar[str] = (
        asyncpg.ClientCannotConnectError.sqlstate
    )
    CONNECTION_DOES_NOT_EXIST: ClassVar[str] = (
        asyncpg.ConnectionDoesNotExistError.sqlstate
    )
    CONNECTION_FAILURE: ClassVar[str] = asyncpg.ConnectionFailureError.sqlstate

    # Constraint violations (400)
    UNIQUE_VIOLATION: ClassVar[str] = asyncpg.UniqueViolationError.sqlstate
    FOREIGN_KEY_VIOLATION: ClassVar[str] = asyncpg.ForeignKeyViolationError.sqlstate
    CHECK_VIOLATION: ClassVar[str] = asyncpg.CheckViolationError.sqlstate

    # Data type and format issues (422)
    INVALID_TEXT_REPRESENTATION: ClassVar[str] = (
        asyncpg.InvalidTextRepresentationError.sqlstate
    )
    INVALID_BINARY_REPRESENTATION: ClassVar[str] = (
        asyncpg.InvalidBinaryRepresentationError.sqlstate
    )
    BAD_COPY_FILE_FORMAT: ClassVar[str] = asyncpg.BadCopyFileFormatError.sqlstate
    UNTRANSLATABLE_CHARACTER: ClassVar[str] = (
        asyncpg.UntranslatableCharacterError.sqlstate
    )

    # Permission and authorization issues (403)
    INSUFFICIENT_PRIVILEGE: ClassVar[str] = asyncpg.InsufficientPrivilegeError.sqlstate
    DEPENDENT_PRIVILEGE_DESCRIPTORS_STILL_EXIST: ClassVar[str] = (
        asyncpg.DependentPrivilegeDescriptorsStillExistError.sqlstate
    )

    # Resource exhaustion (503)
    DISK_FULL: ClassVar[str] = asyncpg.DiskFullError.sqlstate
    OUT_OF_MEMORY: ClassVar[str] = asyncpg.OutOfMemoryError.sqlstate
    TOO_MANY_CONNECTIONS: ClassVar[str] = asyncpg.TooManyConnectionsError.sqlstate

    def __init__(
        self, message: str, status_code: int = HTTPStatus.INTERNAL_SERVER_ERROR
    ) -> None:
        """Initialize with message and HTTP status code.

        Args:
            message: Error message
            status_code: HTTP status code (defaults to 500 Internal Server Error)
        """
        super().__init__(message)
        self.status_code = status_code
