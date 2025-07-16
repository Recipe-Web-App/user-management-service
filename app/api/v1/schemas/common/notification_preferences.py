"""Notification preferences schema definitions."""

from pydantic import BaseModel, Field


class NotificationPreferences(BaseModel):
    """Schema for user notification preferences."""

    email_notifications: bool | None = Field(
        None, description="Whether email notifications are enabled"
    )
    push_notifications: bool | None = Field(
        None, description="Whether push notifications are enabled"
    )
    sms_notifications: bool | None = Field(
        None, description="Whether SMS notifications are enabled"
    )
    marketing_emails: bool | None = Field(
        None, description="Whether marketing emails are enabled"
    )
    security_alerts: bool | None = Field(
        None, description="Whether security alerts are enabled"
    )
    activity_summaries: bool | None = Field(
        None, description="Whether activity summaries are enabled"
    )
    recipe_recommendations: bool | None = Field(
        None, description="Whether recipe recommendations are enabled"
    )
    social_interactions: bool | None = Field(
        None, description="Whether social interaction notifications are enabled"
    )
