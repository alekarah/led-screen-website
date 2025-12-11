"""
Telegram Bot логика для отправки уведомлений
"""

import logging
from telegram import Bot, InlineKeyboardButton, InlineKeyboardMarkup
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

            # Создаем inline клавиатуру с кнопками
            keyboard = None
            if notification.contact_id:
                keyboard = InlineKeyboardMarkup([
                    [
                        InlineKeyboardButton(
                            "✅ Обработано",
                            callback_data=f"processed:{notification.contact_id}"
                        ),
                        InlineKeyboardButton(
                            "🔔 Завтра",
                            callback_data=f"tomorrow:{notification.contact_id}"
                        )
                    ],
                    [
                        InlineKeyboardButton(
                            "👀 Открыть в админке",
                            url="https://s-n-r.ru/admin/contacts"
                        )
                    ]
                ])

            # Отправляем сообщение
            await self.bot.send_message(
                chat_id=self.chat_id,
                text=message,
                parse_mode='HTML',
                reply_markup=keyboard
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

    async def remove_buttons_from_message(self, chat_id: str, message_id: int, success_text: str) -> bool:
        """
        Убрать кнопки из сообщения после обработки и добавить текст о результате

        Args:
            chat_id: ID чата
            message_id: ID сообщения
            success_text: Текст успешного выполнения действия

        Returns:
            bool: True если успешно
        """
        try:
            # Получаем текущее сообщение
            message = await self.bot.edit_message_reply_markup(
                chat_id=chat_id,
                message_id=message_id,
                reply_markup=None  # Убираем кнопки
            )

            # Добавляем текст о результате к оригинальному сообщению
            current_text = message.text if hasattr(message, 'text') else ""
            new_text = f"{current_text}\n\n{success_text}"

            await self.bot.edit_message_text(
                chat_id=chat_id,
                message_id=message_id,
                text=new_text,
                parse_mode='HTML'
            )

            logger.info(f"Кнопки убраны из сообщения {message_id}")
            return True

        except TelegramError as e:
            logger.error(f"Ошибка при редактировании сообщения: {str(e)}")
            return False
