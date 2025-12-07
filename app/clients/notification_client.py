"""HTTP client for notification service integration."""

from http import HTTPStatus

import httpx

from app.api.v1.schemas.downstream.notification import (
    BatchNotificationResponse,
    EmailChangedNotificationRequest,
    NewFollowerNotificationRequest,
)
from app.clients.base_oauth2_service_client import BaseOAuth2ServiceClient
from app.clients.constants import (
    NOTIFICATION_SERVICE_SCOPES,
    get_notification_service_url,
)
from app.core.config import settings
from app.core.logging import get_logger

_log = get_logger(__name__)


class NotificationClient(BaseOAuth2ServiceClient):
    """Client for interacting with the notification service.

    Handles sending notifications to the downstream notification service with automatic
    OAuth2 authentication.
    """

    def __init__(self, environment: str = "local") -> None:
        """Initialize the notification client.

        Args:
            environment: Deployment environment ("local" or "k8s")
        """
        base_url = get_notification_service_url(environment)
        super().__init__(
            base_url=base_url,
            scopes=NOTIFICATION_SERVICE_SCOPES,
            timeout=10.0,  # Shorter timeout for async notification service
        )
        _log.info(f"Initialized NotificationClient for {environment} environment")

    async def send_new_follower_notification(
        self,
        follower_id: str,
        recipient_ids: list[str],
    ) -> BatchNotificationResponse | None:
        """Send new follower notification to recipient(s).

        This method sends a request to the notification service to queue email
        notifications informing users that they have a new follower.

        Args:
            follower_id: UUID of the user who is following
            recipient_ids: List of user IDs who are being followed (max 100)

        Returns:
            BatchNotificationResponse if successful, None if failed (graceful
            degradation)

        Note:
            This method implements graceful degradation - if the notification
            service is unavailable or returns an error, it logs the error but
            does not raise an exception. User operations should not fail due to
            notification failures.
        """
        if not recipient_ids:
            _log.warning(
                "send_new_follower_notification called with empty recipient_ids"
            )
            return None

        request_data = NewFollowerNotificationRequest(
            follower_id=follower_id,
            recipient_ids=recipient_ids,
        )

        try:
            _log.debug(
                f"Sending new follower notification: follower={follower_id}, "
                f"recipients={len(recipient_ids)}"
            )

            response = await self._post(
                endpoint="/notifications/new-follower",
                json_data=request_data.model_dump(),
            )

            # Notification service returns 202 Accepted for queued notifications
            if response.status_code != HTTPStatus.ACCEPTED:
                _log.warning(
                    f"Unexpected status code from notification service: "
                    f"{response.status_code}"
                )
            else:
                batch_response = BatchNotificationResponse(**response.json())
                _log.info(
                    f"Successfully queued {batch_response.queued_count} "
                    f"new follower notification(s)"
                )
                return batch_response

        except httpx.HTTPStatusError as e:
            _log.error(
                f"Notification service HTTP error: {e.response.status_code} - "
                f"{e.response.text}",
                exc_info=True,
            )
            return None

        except httpx.RequestError as e:
            _log.error(
                f"Notification service request failed: {e}",
                exc_info=True,
            )
            return None

        except Exception as e:
            _log.error(
                f"Unexpected error sending notification: {e}",
                exc_info=True,
            )

        return None

    async def send_email_changed_notification(
        self,
        user_id: str,
        old_email: str,
        new_email: str,
    ) -> BatchNotificationResponse | None:
        """Send email change notification to both old and new email addresses.

        This is a security-critical notification that alerts the user when their
        email address is changed. The notification service sends emails to both
        the old and new addresses.

        Args:
            user_id: UUID of the user whose email changed
            old_email: Previous email address
            new_email: New email address

        Returns:
            BatchNotificationResponse if successful, None if failed (graceful
            degradation)

        Note:
            This method implements graceful degradation - if the notification
            service is unavailable or returns an error, it logs the error but
            does not raise an exception. Email updates should not fail due to
            notification failures.
        """
        request_data = EmailChangedNotificationRequest(
            recipient_ids=[user_id],
            old_email=old_email,
            new_email=new_email,
        )

        try:
            _log.debug(
                f"Sending email change notification: user={user_id}, "
                f"old={old_email}, new={new_email}"
            )

            response = await self._post(
                endpoint="/notifications/email-changed",
                json_data=request_data.model_dump(),
            )

            # Notification service returns 202 Accepted for queued notifications
            if response.status_code != HTTPStatus.ACCEPTED:
                _log.warning(
                    f"Unexpected status code from email-changed notification: "
                    f"{response.status_code}"
                )
            else:
                batch_response = BatchNotificationResponse(**response.json())
                _log.info(
                    f"Successfully queued email change notification for user {user_id}"
                )
                return batch_response

        except httpx.HTTPStatusError as e:
            _log.error(
                f"Notification service HTTP error: {e.response.status_code} - "
                f"{e.response.text}",
                exc_info=True,
            )
            return None

        except httpx.RequestError as e:
            _log.error(
                f"Notification service request failed: {e}",
                exc_info=True,
            )
            return None

        except Exception as e:
            _log.error(
                f"Unexpected error sending email change notification: {e}",
                exc_info=True,
            )

        return None


class _NotificationClientSingleton:
    """Singleton holder for notification client instance."""

    _instance: NotificationClient | None = None

    @classmethod
    async def get_instance(cls) -> NotificationClient:
        """Get or create the notification client instance."""
        if cls._instance is None:
            # Determine environment based on settings
            # Use "k8s" for production/kubernetes, otherwise "local"
            env = settings.environment.lower()
            environment = (
                "k8s" if env in ("production", "k8s", "kubernetes") else "local"
            )
            cls._instance = NotificationClient(environment=environment)
            _log.debug(f"Created NotificationClient for {environment}")
        return cls._instance

    @classmethod
    async def close_instance(cls) -> None:
        """Close the notification client instance."""
        if cls._instance is not None:
            await cls._instance.close()
            cls._instance = None
            _log.debug("Closed NotificationClient")


async def get_notification_client() -> NotificationClient:
    """Get or create the global notification client instance.

    Returns:
        NotificationClient instance

    Note:
        This function should be called from FastAPI dependency injection.
        The client will be properly initialized during app startup.
    """
    return await _NotificationClientSingleton.get_instance()


async def close_notification_client() -> None:
    """Close the global notification client.

    This should be called during application shutdown to properly close HTTP
    connections.
    """
    await _NotificationClientSingleton.close_instance()
