"""Base model for Redis entities."""

from datetime import UTC, datetime
from typing import Any

from pydantic import BaseModel, ConfigDict, Field


class BaseRedisModel(BaseModel):
    """Base model for all Redis entities."""

    model_config = ConfigDict(
        json_encoders={
            datetime: lambda v: v.isoformat(),
        },
        validate_assignment=True,
        from_attributes=True,
        frozen=False,
        extra="ignore",
        json_schema_extra={
            "example": {
                "created_at": "2024-01-01T00:00:00+00:00",
                "updated_at": "2024-01-01T00:00:00+00:00",
                "metadata": {"key": "value"},
            }
        },
    )

    created_at: datetime = Field(
        default_factory=lambda: datetime.now(UTC),
        description="Timestamp when the entity was created",
    )
    updated_at: datetime = Field(
        default_factory=lambda: datetime.now(UTC),
        description="Timestamp when the entity was last updated",
    )
    metadata: dict[str, Any] = Field(
        default_factory=dict, description="Additional metadata for the entity"
    )

    def update_timestamp(self) -> None:
        """Update the updated_at timestamp."""
        self.updated_at = datetime.now(UTC)

    def add_metadata(self, key: str, value: Any) -> None:
        """Add metadata to the model."""
        self.metadata[key] = value
        self.update_timestamp()

    def get_metadata(self, key: str, default: Any = None) -> Any:
        """Get metadata from the model."""
        return self.metadata.get(key, default)
