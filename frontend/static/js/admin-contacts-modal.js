// Модалка «Подробнее» для страницы ЗАЯВОК (inbox).
// Показывает инфо, меняет статус, а заметки/напоминание делегирует в ContactsNotes.

(function (w) {
  function initModal() {
    const overlay = document.getElementById('contact-details-modal');
    if (!overlay) return;

    const closeBtn = overlay.querySelector('.modal-close');

    const f = {
      // вкладки (DOM в шаблоне уже есть)
      tabInfoBtn:  document.getElementById('tab-info-btn'),
      tabNotesBtn: document.getElementById('tab-notes-btn'),
      tabInfo:     document.getElementById('cd-tab-info'),
      tabNotes:    document.getElementById('cd-tab-notes'),

      // инфо
      name:    document.getElementById('cd-name'),
      phone:   document.getElementById('cd-phone'),
      email:   document.getElementById('cd-email'),
      company: document.getElementById('cd-company'),
      type:    document.getElementById('cd-type'),
      date:    document.getElementById('cd-date'),
      status:  document.getElementById('cd-status'),
      msg:     document.getElementById('cd-message'),

      // действия (инбокс)
      btnProcessed: document.getElementById('cd-mark-processed'),
      btnNew:       document.getElementById('cd-mark-new'),
      btnArchive:   document.getElementById('cd-archive'),

      // напоминание
      remAt:   document.getElementById('cd-remind-at'),
      remSave: document.getElementById('cd-reminder-save'),
      remClear:document.getElementById('cd-reminder-clear'),

      // заметки
      notesList:  document.getElementById('cd-notes-list'),
      noteAuthor: document.getElementById('cd-note-author'),
      noteText:   document.getElementById('cd-note-text'),
      noteAdd:    document.getElementById('cd-note-add'),
    };

    const hasReminder = !!(f.remAt && f.remSave && f.remClear);
    let currentId = null;

    // Инициализируем общий модуль заметок/напоминания
    const Notes = w.ContactsNotes?.init({
      getCurrentId: () => currentId,
      els: {
        tabInfoBtn: f.tabInfoBtn,
        tabNotesBtn: f.tabNotesBtn,
        tabInfo:     f.tabInfo,
        tabNotes:    f.tabNotes,
        remAt:    hasReminder ? f.remAt    : null,
        remSave:  hasReminder ? f.remSave  : null,
        remClear: hasReminder ? f.remClear : null,
        notesList:  f.notesList,
        noteAuthor: f.noteAuthor,
        noteText:   f.noteText,
        noteAdd:    f.noteAdd,
      },
    });

    function open()  { overlay.classList.remove('hidden'); overlay.setAttribute('aria-hidden','false'); }
    function close() { overlay.classList.add('hidden');    overlay.setAttribute('aria-hidden','true');  }

    closeBtn?.addEventListener('click', close);
    overlay?.addEventListener('click', (e) => { if (e.target === overlay) close(); });

    // Открытие модалки по кнопке «Подробнее»
    document.body.addEventListener('click', async (e) => {
      const btn = e.target.closest('.js-contact-details');
      if (!btn) return;

      currentId = Number(btn.dataset.id);

      // инфо
      f.name.textContent = btn.dataset.name || '—';

      const ph = btn.dataset.phone || '';
      f.phone.textContent = ph || '—';
      f.phone.href = ph ? `tel:${ph}` : '#';

      const em = btn.dataset.email || '';
      f.email.textContent = em || '—';
      f.email.href = em ? `mailto:${em}` : '#';

      f.company.textContent = btn.dataset.company || '—';
      f.type.textContent    = btn.dataset.type || '—';
      f.date.textContent    = btn.dataset.date || '—';
      f.msg.textContent     = btn.dataset.message || '';

      w.ContactsUI.setModalStatusBadge(f.status, btn.dataset.status || 'new');

      // вкладка «Заметки»: подставить напоминание и загрузить заметки
      const rem = btn.dataset.remindAt || '';
      if (hasReminder) {
        await Notes?.onOpen({ remindAt: rem });
      } else {
        await Notes?.onOpen({}); // только заметки, без поля remind_at
      }

      open();
    });

    // Смена статуса из модалки
    async function change(status, okMsg) {
      if (!currentId) return;
      try {
        await w.ContactsAPI.updateStatus(currentId, status);
        w.ContactsUI.setModalStatusBadge(f.status, status);
        w.ContactsUI.setRowStatusById(currentId, status);
        w.ContactsUI.show('ok', okMsg);
      } catch (err) {
        w.ContactsUI.show('error', err.message);
      }
    }

    f.btnProcessed?.addEventListener('click', () => change('processed', 'Помечено как обработано'));
    f.btnNew?.addEventListener('click',       () => change('new',       'Возвращено в новые'));
    f.btnArchive?.addEventListener('click',   () => change('archived',  'Отправлено в архив'));
  }

  w.ContactsModalInit = initModal;
})(window);
