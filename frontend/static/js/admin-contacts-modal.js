// Модалка «Подробнее»: открытие/закрытие и смена статуса
(function (w) {
    function initModal() {
        const overlay = document.getElementById('contact-details-modal');
        if (!overlay) return; // если модалка на странице не отрисована

        const modal = overlay.querySelector('.modal');
        const closeBtn = overlay.querySelector('.modal-close');

        const f = {
        name:    document.getElementById('cd-name'),
        phone:   document.getElementById('cd-phone'),
        email:   document.getElementById('cd-email'),
        company: document.getElementById('cd-company'),
        type:    document.getElementById('cd-type'),
        date:    document.getElementById('cd-date'),
        status:  document.getElementById('cd-status'),
        msg:     document.getElementById('cd-message'),
        btnProcessed: document.getElementById('cd-mark-processed'),
        btnNew:       document.getElementById('cd-mark-new'),
        btnArchive:   document.getElementById('cd-archive'),
        };

        let currentId = null;

        function open()  { overlay.classList.remove('hidden'); overlay.setAttribute('aria-hidden','false'); }
        function close() { overlay.classList.add('hidden');    overlay.setAttribute('aria-hidden','true');  }

        closeBtn?.addEventListener('click', close);
        overlay?.addEventListener('click', (e) => { if (e.target === overlay) close(); });

        // открыть модалку по кнопке «Подробнее»
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
        f.msg.textContent     = btn.dataset.message || '';

        window.ContactsUI.setModalStatusBadge(f.status, btn.dataset.status || 'new');

        open();
        });

        // смена статуса из модалки
        async function change(status, okMsg) {
        if (!currentId) return;
        try {
            await window.ContactsAPI.updateStatus(currentId, status);
            window.ContactsUI.setModalStatusBadge(f.status, status);
            window.ContactsUI.setRowStatusById(currentId, status);
            window.ContactsUI.show('ok', okMsg);
        } catch (err) {
            window.ContactsUI.show('error', err.message);
        }
        }

        f.btnProcessed?.addEventListener('click', () => change('processed', 'Помечено как обработано'));
        f.btnNew?.addEventListener('click',       () => change('new',       'Возвращено в новые'));
        f.btnArchive?.addEventListener('click',   () => change('archived',  'Отправлено в архив'));
    }

    w.ContactsModalInit = initModal;
})(window);
