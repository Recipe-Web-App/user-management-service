"""Notification preference keys."""

from enum import Enum


class NotificationPreferenceKey(str, Enum):
    """Notification preference keys."""

    EMAIL_NOTIFICATIONS = "email_notifications"
    PUSH_NOTIFICATIONS = "push_notifications"
    SMS_NOTIFICATIONS = "sms_notifications"
    MARKETING_EMAILS = "marketing_emails"
    SECURITY_ALERTS = "security_alerts"
    ACTIVITY_SUMMARIES = "activity_summaries"
    RECIPE_RECOMMENDATIONS = "recipe_recommendations"
    SOCIAL_INTERACTIONS = "social_interactions"
