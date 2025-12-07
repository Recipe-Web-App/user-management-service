"""Difficulty level enum for recipes."""

import enum


class DifficultyLevelEnum(str, enum.Enum):
    """Enum for recipe difficulty levels."""

    BEGINNER = "BEGINNER"
    EASY = "EASY"
    MEDIUM = "MEDIUM"
    HARD = "HARD"
    EXPERT = "EXPERT"
