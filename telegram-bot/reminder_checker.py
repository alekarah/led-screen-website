"""
Фоновая задача для проверки и отправки напоминаний
"""

import logging
import asyncio
from telegram import Bot
from telegram.error import TelegramError
import httpx

logger = logging.getLogger(__name__)


class ReminderChecker:
    """Проверяет напоминания и отправляет уведомления"""

    def __init__(self, bot: Bot, chat_id: str, backend_url: str, check_interval: int = 300):
        """
        Инициализация проверщика напоминаний

        Args:
            bot: Telegram Bot instance
            chat_id: ID чата для отправки напоминаний
            backend_url: URL Go backend API
            check_interval: Интервал проверки в секундах (по умолчанию 300 = 5 минут)
        """
        self.bot = bot
        self.chat_id = chat_id
        self.backend_url = backend_url
        self.check_interval = check_interval
        self._task = None
        logger.info(f"ReminderChecker инициализирован (проверка каждые {check_interval} сек)")

    async def start(self):
        """Запустить фоновую задачу проверки напоминаний"""
        if self._task and not self._task.done():
            logger.warning("ReminderChecker уже запущен")
            return

        logger.info("Запуск фоновой задачи проверки напоминаний...")
        self._task = asyncio.create_task(self._check_loop())

    async def stop(self):
        """Остановить фоновую задачу"""
        if self._task and not self._task.done():
            logger.info("Остановка проверки напоминаний...")
            self._task.cancel()
            try:
                await self._task
            except asyncio.CancelledError:
                pass
            logger.info("✓ Проверка напоминаний остановлена")

    async def _check_loop(self):
        """Основной цикл проверки напоминаний"""
        while True:
            try:
                await self._check_and_send_reminders()
            except Exception as e:
                logger.error(f"Ошибка при проверке напоминаний: {e}")

            # Ждём до следующей проверки
            await asyncio.sleep(self.check_interval)

    async def _check_and_send_reminders(self):
        """Проверить напоминания и отправить уведомления"""
        try:
            async with httpx.AsyncClient() as client:
                # Получаем список напоминаний которые пора отправить
                response = await client.get(
                    f"{self.backend_url}/api/telegram/due-reminders",
                    timeout=10.0
                )
                response.raise_for_status()

                data = response.json()
                reminders = data.get("reminders", [])

                if not reminders:
                    logger.debug("Нет напоминаний для отправки")
                    return

                logger.info(f"Найдено {len(reminders)} напоминаний для отправки")

                # Отправляем уведомление для каждого напоминания
                for reminder in reminders:
                    try:
                        await self._send_reminder_notification(reminder)

                        # Помечаем напоминание как отправленное
                        await client.post(
                            f"{self.backend_url}/api/telegram/mark-reminder-sent",
                            json={"contact_id": reminder["contact_id"]},
                            timeout=10.0
                        )

                        logger.info(f"✓ Напоминание отправлено для контакта ID {reminder['contact_id']}")

                    except Exception as e:
                        logger.error(f"Ошибка при отправке напоминания для ID {reminder['contact_id']}: {e}")

        except httpx.HTTPError as e:
            logger.error(f"Ошибка HTTP при получении напоминаний: {e}")
        except Exception as e:
            logger.error(f"Неожиданная ошибка при проверке напоминаний: {e}")

    async def _send_reminder_notification(self, reminder: dict):
        """
        Отправить уведомление о напоминании в Telegram

        Args:
            reminder: Данные напоминания
        """
        try:
            # Формируем сообщение
            message_parts = [
                "⏰ <b>Напоминание!</b>",
                "",
                f"Пора связаться с клиентом:",
                f"👤 <b>Имя:</b> {reminder['name']}",
                f"📞 <b>Телефон:</b> {reminder['phone']}"
            ]

            if reminder.get('email'):
                message_parts.append(f"📧 <b>Email:</b> {reminder['email']}")

            if reminder.get('company'):
                message_parts.append(f"🏢 <b>Компания:</b> {reminder['company']}")

            if reminder.get('project_type'):
                message_parts.append(f"📋 <b>Тип проекта:</b> {reminder['project_type']}")

            message_parts.append("")
            message_parts.append(f"🕐 <b>Запланировано на:</b> {reminder['remind_at']}")

            # Добавляем ссылку на админку
            admin_url = f"https://s-n-r.ru/admin/contacts"
            message_parts.append("")
            message_parts.append(f"<a href='{admin_url}'>📊 Открыть в админке</a>")

            message = "\n".join(message_parts)

            # Отправляем сообщение
            await self.bot.send_message(
                chat_id=self.chat_id,
                text=message,
                parse_mode='HTML'
            )

        except TelegramError as e:
            logger.error(f"Ошибка Telegram API при отправке напоминания: {e}")
            raise
