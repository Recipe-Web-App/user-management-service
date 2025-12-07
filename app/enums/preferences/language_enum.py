"""Language enumeration for user preferences."""

from enum import Enum


class LanguageEnum(str, Enum):
    """Enum for supported language preferences."""

    EN = "EN"
    ES = "ES"
    FR = "FR"
    DE = "DE"
    IT = "IT"
    PT = "PT"
    ZH = "ZH"
    JA = "JA"
    KO = "KO"
    RU = "RU"
