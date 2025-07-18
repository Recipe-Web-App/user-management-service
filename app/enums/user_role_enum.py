"""User role enum for distinguishing admin and regular users."""

from enum import Enum


class UserRoleEnum(str, Enum):
    """Enumeration of user roles."""

    ADMIN = "ADMIN"
    USER = "USER"
