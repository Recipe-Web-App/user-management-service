"""Notification preferences schema definitions."""

from pydantic import BaseModel, Field


class NotificationPreferences(BaseModel):
    """Schema for user notification preferences."""

    email_notifications: bool = Field(
        ..., description="Whether email notifications are enabled"
    )
    push_notifications: bool = Field(
        ..., description="Whether push notifications are enabled"
    )
    sms_notifications: bool = Field(
        ..., description="Whether SMS notifications are enabled"
    )
    marketing_emails: bool = Field(
        ..., description="Whether marketing emails are enabled"
    )
    security_alerts: bool = Field(
        ..., description="Whether security alerts are enabled"
    )
    activity_summaries: bool = Field(
        ..., description="Whether activity summaries are enabled"
    )
    recipe_recommendations: bool = Field(
        ..., description="Whether recipe recommendations are enabled"
    )
    social_interactions: bool = Field(
        ..., description="Whether social interaction notifications are enabled"
    )
