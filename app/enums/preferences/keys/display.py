"""Display preference keys."""

from enum import Enum


class DisplayPreferenceKey(str, Enum):
    """Display preference keys."""

    FONT_SIZE = "font_size"
    COLOR_SCHEME = "color_scheme"
    LAYOUT_DENSITY = "layout_density"
    SHOW_IMAGES = "show_images"
    COMPACT_MODE = "compact_mode"
