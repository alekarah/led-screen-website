(function (w) {
  function initArchiveModal() {
    const overlay = document.getElementById('contact-details-modal');
    if (!overlay) return;

    const closeBtn = overlay.querySelector('.modal-close');
    const f = {
      name:    document.getElementById('cd-name'),
      phone:   document.getElementById('cd-phone'),
      email:   document.getElementById('cd-email'),
      company: document.getElementById('cd-company'),
      type:    document.getElementById('cd-type'),
      date:    document.getElementById('cd-date'),
      archived:document.getElementById('cd-archived'),
      msg:     document.getElementById('cd-message'),
      btnRestore: document.getElementById('cd-restore'),
      btnDelete:  document.getElementById('cd-delete'),
    };

    let currentId = null;

    function open()  { overlay.classList.remove('hidden'); overlay.setAttribute('aria-hidden','false'); }
    function close() { overlay.classList.add('hidden');    overlay.setAttribute('aria-hidden','true');  }

    closeBtn?.addEventListener('click', close);
    overlay?.addEventListener('click', (e) => { if (e.target === overlay) close(); });

    document.body.addEventListener('click', (e) => {
      const btn = e.target.closest('.js-contact-details');
      if (!btn) return;

      currentId = Number(btn.dataset.id);

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
      f.archived.textContent= btn.dataset.archived || '—';
      f.msg.textContent     = btn.dataset.message || '';

      w.ContactsUI.setModalStatusBadge(f.status, 'archived');

      open();
    });

    f.btnRestore?.addEventListener('click', async () => {
      if (!currentId) return;
      try {
        await w.ContactsAPI.restore(currentId, 'new');
        w.ContactsUI.removeRowById(currentId);
        w.ContactsUI.show('ok', 'Восстановлено');
        close();
      } catch (err) { w.ContactsUI.show('error', err.message); }
    });

    f.btnDelete?.addEventListener('click', async () => {
      if (!currentId) return;
      if (!confirm('Удалить заявку безвозвратно?')) return;
      try {
        await w.ContactsAPI.remove(currentId, { hard: true });
        w.ContactsUI.removeRowById(currentId);
        w.ContactsUI.show('ok', 'Удалено');
        close();
      } catch (err) { w.ContactsUI.show('error', err.message); }
    });
  }

  w.ContactsArchiveModalInit = initArchiveModal;
})(window);
