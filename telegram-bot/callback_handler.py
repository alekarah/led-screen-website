"""
Обработчик callback queries от inline кнопок в Telegram
"""

import logging
from datetime import datetime, timedelta
from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import ContextTypes
import httpx
from config import settings

logger = logging.getLogger(__name__)


class CallbackHandler:
    """Обработчик нажатий на inline кнопки"""

    def __init__(self, backend_url: str):
        """
        Инициализация обработчика

        Args:
            backend_url: URL Go backend API
        """
        self.backend_url = backend_url
        logger.info(f"CallbackHandler инициализирован. Backend URL: {backend_url}")

    async def handle_callback_query(self, update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
        """
        Обработка callback query от inline кнопок

        Args:
            update: Telegram Update объект
            context: Контекст бота
        """
        query = update.callback_query
        await query.answer()  # Убираем "loading" на кнопке

        # Парсим callback_data: "action:contact_id"
        callback_data = query.data
        if not callback_data or ':' not in callback_data:
            logger.error(f"Неверный формат callback_data: {callback_data}")
            await query.edit_message_text(
                text=query.message.text + "\n\n❌ Ошибка: неверный формат данных",
                parse_mode='HTML'
            )
            return

        action, contact_id_str = callback_data.split(':', 1)

        try:
            contact_id = int(contact_id_str)
        except ValueError:
            logger.error(f"Неверный ID контакта: {contact_id_str}")
            await query.edit_message_text(
                text=query.message.text + "\n\n❌ Ошибка: неверный ID контакта",
                parse_mode='HTML'
            )
            return

        # Обрабатываем разные действия
        if action == "processed":
            await self._handle_processed(query, contact_id)
        elif action == "tomorrow":
            await self._handle_tomorrow(query, contact_id)
        else:
            logger.error(f"Неизвестное действие: {action}")
            await query.edit_message_text(
                text=query.message.text + "\n\n❌ Ошибка: неизвестное действие",
                parse_mode='HTML'
            )

    async def _handle_processed(self, query, contact_id: int) -> None:
        """
        Обработка нажатия кнопки "Обработано"

        Args:
            query: CallbackQuery объект
            contact_id: ID контакта
        """
        try:
            async with httpx.AsyncClient() as client:
                # Меняем статус на "processed"
                status_response = await client.post(
                    f"{self.backend_url}/api/telegram/update-status",
                    json={"contact_id": contact_id, "status": "processed"},
                    timeout=10.0
                )
                status_response.raise_for_status()

                # Добавляем системную заметку
                note_response = await client.post(
                    f"{self.backend_url}/api/telegram/add-note",
                    json={
                        "contact_id": contact_id,
                        "text": "Обработано",
                        "author": "Telegram Bot"
                    },
                    timeout=10.0
                )
                note_response.raise_for_status()

                # Оставляем только кнопку "Открыть в админке"
                keyboard = InlineKeyboardMarkup([
                    [
                        InlineKeyboardButton(
                            "👀 Открыть в админке",
                            url="https://s-n-r.ru/admin/contacts"
                        )
                    ]
                ])

                success_text = "✅ <b>Обработано</b>"
                await query.edit_message_text(
                    text=query.message.text + "\n\n" + success_text,
                    parse_mode='HTML',
                    reply_markup=keyboard
                )

                logger.info(f"Контакт {contact_id} помечен как обработанный через Telegram")

        except httpx.HTTPStatusError as e:
            logger.error(f"Ошибка HTTP при обработке: {e}")
            await query.edit_message_text(
                text=query.message.text + "\n\n❌ Ошибка при обработке заявки",
                parse_mode='HTML'
            )
        except Exception as e:
            logger.error(f"Ошибка при обработке 'processed': {e}")
            await query.edit_message_text(
                text=query.message.text + "\n\n❌ Неожиданная ошибка",
                parse_mode='HTML'
            )

    async def _handle_tomorrow(self, query, contact_id: int) -> None:
        """
        Обработка нажатия кнопки "Завтра"

        Args:
            query: CallbackQuery объект
            contact_id: ID контакта
        """
        try:
            # Вычисляем время напоминания: завтра в 9:00
            tomorrow = datetime.now() + timedelta(days=1)
            remind_at = tomorrow.replace(hour=9, minute=0, second=0, microsecond=0)
            remind_at_str = remind_at.strftime("%Y-%m-%d %H:%M")

            async with httpx.AsyncClient() as client:
                # Устанавливаем напоминание
                response = await client.post(
                    f"{self.backend_url}/api/telegram/set-reminder",
                    json={
                        "contact_id": contact_id,
                        "remind_at": remind_at_str
                    },
                    timeout=10.0
                )
                response.raise_for_status()

                # Оставляем только кнопку "Открыть в админке"
                keyboard = InlineKeyboardMarkup([
                    [
                        InlineKeyboardButton(
                            "👀 Открыть в админке",
                            url="https://s-n-r.ru/admin/contacts"
                        )
                    ]
                ])

                success_text = f"🔔 <b>Напоминание установлено на {remind_at.strftime('%d.%m.%Y в 09:00')}</b>"
                await query.edit_message_text(
                    text=query.message.text + "\n\n" + success_text,
                    parse_mode='HTML',
                    reply_markup=keyboard
                )

                logger.info(f"Напоминание для контакта {contact_id} установлено на {remind_at_str}")

        except httpx.HTTPStatusError as e:
            logger.error(f"Ошибка HTTP при установке напоминания: {e}")
            await query.edit_message_text(
                text=query.message.text + "\n\n❌ Ошибка при установке напоминания",
                parse_mode='HTML'
            )
        except Exception as e:
            logger.error(f"Ошибка при обработке 'tomorrow': {e}")
            await query.edit_message_text(
                text=query.message.text + "\n\n❌ Неожиданная ошибка",
                parse_mode='HTML'
            )
