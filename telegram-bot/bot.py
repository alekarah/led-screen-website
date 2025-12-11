"""
Telegram Bot логика для отправки уведомлений
"""

import logging
from telegram import Bot
from telegram.error import TelegramError
from typing import Optional

logger = logging.getLogger(__name__)


class TelegramNotifier:
    """Класс для отправки уведомлений в Telegram"""

    def __init__(self, bot_token: str, chat_id: str):
        """
        Инициализация бота

        Args:
            bot_token: Токен Telegram бота
            chat_id: ID чата для отправки сообщений
        """
        self.bot = Bot(token=bot_token)
        self.chat_id = chat_id
        logger.info(f"Telegram бот инициализирован. Chat ID: {chat_id}")

    async def send_new_contact_notification(self, notification) -> bool:
        """
        Отправить уведомление о новой заявке

        Args:
            notification: Объект ContactNotification с данными заявки

        Returns:
            bool: True если отправлено успешно, False в противном случае
        """
        try:
            # Формируем красивое сообщение
            message = self._format_new_contact_message(notification)

            # Отправляем сообщение
            await self.bot.send_message(
                chat_id=self.chat_id,
                text=message,
                parse_mode='HTML'
            )

            logger.info(f"Уведомление отправлено в Telegram для: {notification.name}")
            return True

        except TelegramError as e:
            logger.error(f"Ошибка Telegram API: {str(e)}")
            return False
        except Exception as e:
            logger.error(f"Неожиданная ошибка при отправке: {str(e)}")
            return False

    def _format_new_contact_message(self, notification) -> str:
        """
        Форматировать сообщение о новой заявке

        Args:
            notification: Данные заявки

        Returns:
            str: Отформатированное сообщение
        """
        message_parts = [
            "🆕 <b>Новая заявка с сайта!</b>",
            "",
            f"👤 <b>Имя:</b> {notification.name}",
            f"📞 <b>Телефон:</b> {notification.phone}"
        ]

        # Добавляем опциональные поля
        if notification.email:
            message_parts.append(f"📧 <b>Email:</b> {notification.email}")

        if notification.company:
            message_parts.append(f"🏢 <b>Компания:</b> {notification.company}")

        if notification.project_type:
            message_parts.append(f"📋 <b>Тип проекта:</b> {notification.project_type}")

        if notification.message:
            # Обрезаем длинные сообщения
            msg_text = notification.message
            if len(msg_text) > 200:
                msg_text = msg_text[:200] + "..."
            message_parts.append(f"💬 <b>Сообщение:</b> {msg_text}")

        if notification.timestamp:
            message_parts.append(f"🕐 <b>Получена:</b> {notification.timestamp}")

        # Добавляем ссылку на админку
        if notification.contact_id:
            admin_url = f"https://s-n-r.ru/admin/contacts"
            message_parts.append("")
            message_parts.append(f"<a href='{admin_url}'>📊 Открыть в админке</a>")

        return "\n".join(message_parts)

    async def send_reminder_notification(self, contact_name: str, phone: str, note: str) -> bool:
        """
        Отправить напоминание о перезвоне

        Args:
            contact_name: Имя контакта
            phone: Телефон
            note: Текст заметки

        Returns:
            bool: True если отправлено успешно
        """
        try:
            message = (
                "⏰ <b>Напоминание!</b>\n\n"
                f"Пора связаться с клиентом:\n"
                f"👤 {contact_name}\n"
                f"📞 {phone}\n\n"
                f"📝 <b>Заметка:</b> {note}"
            )

            await self.bot.send_message(
                chat_id=self.chat_id,
                text=message,
                parse_mode='HTML'
            )

            logger.info(f"Напоминание отправлено для: {contact_name}")
            return True

        except TelegramError as e:
            logger.error(f"Ошибка при отправке напоминания: {str(e)}")
            return False
