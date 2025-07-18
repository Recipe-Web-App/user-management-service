"""Environment enum for system health."""

from enum import Enum


class EnvironmentEnum(str, Enum):
    """Environment types for the application."""

    DEVELOPMENT = "development"
    STAGING = "staging"
    PRODUCTION = "production"
