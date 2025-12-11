"""
LED Screen Website - Telegram Notification Bot
FastAPI приложение для отправки уведомлений в Telegram
"""

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Optional
import logging
import asyncio
from telegram.ext import Application, CallbackQueryHandler
from bot import TelegramNotifier
from callback_handler import CallbackHandler
from config import settings

# Настройка логирования
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Инициализация FastAPI
app = FastAPI(
    title="LED Screen Telegram Bot",
    description="Сервис для отправки уведомлений о новых заявках в Telegram",
    version="1.0.0"
)

# Инициализация Telegram бота для отправки уведомлений
notifier = TelegramNotifier(
    bot_token=settings.TELEGRAM_BOT_TOKEN,
    chat_id=settings.TELEGRAM_CHAT_ID
)

# Инициализация обработчика callback queries
# Backend URL - localhost т.к. Go и Python работают на одном сервере
callback_handler = CallbackHandler(backend_url="http://127.0.0.1:8080")

# Глобальная переменная для Telegram Application
telegram_app = None


class ContactNotification(BaseModel):
    """Модель данных для уведомления о новой заявке"""
    name: str
    phone: str
    email: Optional[str] = None
    company: Optional[str] = None
    project_type: Optional[str] = None
    message: Optional[str] = None
    contact_id: Optional[int] = None
    timestamp: Optional[str] = None


@app.on_event("startup")
async def startup_event():
    """Запуск Telegram бота при старте FastAPI"""
    global telegram_app

    logger.info("Инициализация Telegram Application...")

    # Создаем Application
    telegram_app = Application.builder().token(settings.TELEGRAM_BOT_TOKEN).build()

    # Добавляем обработчик callback queries
    telegram_app.add_handler(
        CallbackQueryHandler(callback_handler.handle_callback_query)
    )

    # Запускаем бота в фоне
    await telegram_app.initialize()
    await telegram_app.start()
    await telegram_app.updater.start_polling()

    logger.info("✓ Telegram bot запущен и слушает callback queries")


@app.on_event("shutdown")
async def shutdown_event():
    """Остановка Telegram бота при остановке FastAPI"""
    global telegram_app

    if telegram_app:
        logger.info("Остановка Telegram bot...")
        await telegram_app.updater.stop()
        await telegram_app.stop()
        await telegram_app.shutdown()
        logger.info("✓ Telegram bot остановлен")


@app.get("/")
async def root():
    """Healthcheck endpoint"""
    return {
        "status": "running",
        "service": "LED Screen Telegram Notification Bot",
        "version": "1.0.0"
    }


@app.get("/health")
async def health_check():
    """Проверка здоровья сервиса"""
    return {"status": "healthy"}


@app.post("/api/send-notification")
async def send_notification(notification: ContactNotification):
    """
    Отправить уведомление о новой заявке в Telegram

    Args:
        notification: Данные о заявке

    Returns:
        dict: Статус отправки
    """
    try:
        logger.info(f"Получен запрос на отправку уведомления для: {notification.name}")

        # Отправляем уведомление
        success = await notifier.send_new_contact_notification(notification)

        if success:
            logger.info(f"Уведомление успешно отправлено для: {notification.name}")
            return {
                "status": "success",
                "message": "Notification sent successfully"
            }
        else:
            logger.error(f"Не удалось отправить уведомление для: {notification.name}")
            raise HTTPException(status_code=500, detail="Failed to send notification")

    except Exception as e:
        logger.error(f"Ошибка при отправке уведомления: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app",
        host="127.0.0.1",
        port=5000,
        reload=False,
        log_level="info"
    )
