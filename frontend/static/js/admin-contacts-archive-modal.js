// Модалка «Подробнее» для страницы АРХИВА.
// Информацию/кнопки обрабатываем здесь, а заметки/напоминания — через общий ContactsNotes.

(function (w) {
  function initArchiveModal() {
    const overlay = document.getElementById('contact-details-modal');
    if (!overlay) return;

    const closeBtn = overlay.querySelector('.modal-close');

    const f = {
      // вкладки (DOM уже есть в шаблоне)
      tabInfoBtn:  document.getElementById('tab-info-btn'),
      tabNotesBtn: document.getElementById('tab-notes-btn'),
      tabInfo:     document.getElementById('cd-tab-info'),
      tabNotes:    document.getElementById('cd-tab-notes'),

      // инфо
      name:     document.getElementById('cd-name'),
      phone:    document.getElementById('cd-phone'),
      email:    document.getElementById('cd-email'),
      company:  document.getElementById('cd-company'),
      type:     document.getElementById('cd-type'),
      date:     document.getElementById('cd-date'),
      archived: document.getElementById('cd-archived'),
      msg:      document.getElementById('cd-message'),

      // кнопки архива
      btnRestore: document.getElementById('cd-restore'),
      btnDelete:  document.getElementById('cd-delete'),

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

    let currentId = null;

    // Общий модуль «Заметки + Перезвонить позже»
    const Notes = w.ContactsNotes?.init({
      getCurrentId: () => currentId,
      els: {
        tabInfoBtn: f.tabInfoBtn,
        tabNotesBtn: f.tabNotesBtn,
        tabInfo:     f.tabInfo,
        tabNotes:    f.tabNotes,
        remAt:   f.remAt,
        remSave: f.remSave,
        remClear:f.remClear,
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

    // Открытие модалки
    document.body.addEventListener('click', async (e) => {
      const btn = e.target.closest('.js-contact-details');
      if (!btn) return;

      currentId = Number(btn.dataset.id);

      f.name.textContent     = btn.dataset.name || '—';

      const ph = btn.dataset.phone || '';
      f.phone.textContent = ph || '—';
      f.phone.href = ph ? `tel:${ph}` : '#';

      const em = btn.dataset.email || '';
      f.email.textContent = em || '—';
      f.email.href = em ? `mailto:${em}` : '#';

      f.company.textContent  = btn.dataset.company || '—';
      f.type.textContent     = btn.dataset.type || '—';
      f.date.textContent     = btn.dataset.date || '—';
      f.archived.textContent = btn.dataset.archived || '—';
      f.msg.textContent      = btn.dataset.message || '';

      // вкладка «Заметки»: подставить напоминание и загрузить заметки
      const rem = btn.dataset.remindAt || '';
      await Notes?.onOpen({ remindAt: rem });

      open();
    });

    // Восстановить из архива
    f.btnRestore?.addEventListener('click', async () => {
      if (!currentId) return;
      try {
        await w.ContactsAPI.restore(currentId, 'new');
        w.ContactsUI.removeRowById(currentId);
        w.ContactsUI.show('ok', 'Восстановлено');
        close();
      } catch (err) {
        w.ContactsUI.show('error', err.message);
      }
    });

    // Удалить навсегда
    f.btnDelete?.addEventListener('click', async () => {
      if (!currentId) return;
      if (!confirm('Удалить заявку безвозвратно?')) return;
      try {
        await w.ContactsAPI.remove(currentId, { hard: true });
        w.ContactsUI.removeRowById(currentId);
        w.ContactsUI.show('ok', 'Удалено');
        close();
      } catch (err) {
        w.ContactsUI.show('error', err.message);
      }
    });
  }

  w.ContactsArchiveModalInit = initArchiveModal;
})(window);
