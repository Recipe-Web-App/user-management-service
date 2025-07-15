"""Privacy preference keys."""

from enum import Enum


class PrivacyPreferenceKey(str, Enum):
    """Privacy preference keys."""

    PROFILE_VISIBILITY = "profile_visibility"
    RECIPE_VISIBILITY = "recipe_visibility"
    ACTIVITY_VISIBILITY = "activity_visibility"
    CONTACT_INFO_VISIBILITY = "contact_info_visibility"
    DATA_SHARING = "data_sharing"
    ANALYTICS_TRACKING = "analytics_tracking"
