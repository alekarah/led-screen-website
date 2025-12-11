"""
Конфигурация для Telegram бота
"""

import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Настройки приложения"""

    # Telegram настройки
    TELEGRAM_BOT_TOKEN: str
    TELEGRAM_CHAT_ID: str

    # Настройки сервера
    HOST: str = "127.0.0.1"
    PORT: int = 5000

    # Режим работы
    ENVIRONMENT: str = "production"

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"


# Создаем экземпляр настроек
settings = Settings()
