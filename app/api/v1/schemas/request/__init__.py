"""Request schemas for the API."""

from . import notification, preference, user

__all__ = notification.__all__ + preference.__all__ + user.__all__
