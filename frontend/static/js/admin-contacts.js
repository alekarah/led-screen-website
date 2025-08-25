document.addEventListener('DOMContentLoaded', () => {
    // ---------- Делегирование: "Обработать" ----------
    document.addEventListener('click', async (e) => {
        const btn = e.target.closest('.mark-done');
        if (!btn) return;

        const tr = btn.closest('tr');
        const id = tr?.dataset?.id;
        if (!id) return;

        btn.disabled = true;

        try {
        const res = await fetch(`/admin/contacts/${id}/status`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status: 'processed' })
        });
        let data = {};
        try { data = await res.json(); } catch (_) {}

        if (!res.ok) throw new Error(data.error || 'Не удалось пометить как обработано');

        const cell = tr.querySelector('.status-cell');
        if (cell) cell.innerHTML = '<span class="badge badge-ok">Обработано</span>';

        const detailsBtn = tr.querySelector('.js-contact-details');
        if (detailsBtn) detailsBtn.dataset.status = 'processed';
        } catch (err) {
        console.error(err);
        if (window.showAdminMessage) showAdminMessage(err.message, 'error');
        } finally {
        btn.disabled = false;
        }
    });

    // ---------- Модалка "Подробнее" ----------
    const overlay = document.getElementById('contact-details-modal');
    if (!overlay) return;

    const closeBtn = overlay.querySelector('.modal-close');
    const f = {
        name: document.getElementById('cd-name'),
        phone: document.getElementById('cd-phone'),
        email: document.getElementById('cd-email'),
        company: document.getElementById('cd-company'),
        type: document.getElementById('cd-type'),
        date: document.getElementById('cd-date'),
        status: document.getElementById('cd-status'),
        message: document.getElementById('cd-message'),
    };

    let lastFocused = null;
    let currentContactId = null;

    function openModal() {
        lastFocused = document.activeElement;
        document.body.classList.add('modal-open');
        overlay.classList.remove('hidden');
        overlay.setAttribute('aria-hidden', 'false');
        closeBtn?.focus(); // удобнее закрыть сразу
    }

    function closeModal() {
        overlay.classList.add('hidden');
        overlay.setAttribute('aria-hidden', 'true');
        document.body.classList.remove('modal-open');
        lastFocused?.focus?.();
    }

    // закрытия
    if (closeBtn) closeBtn.addEventListener('click', closeModal);
    overlay.addEventListener('click', (e) => { if (e.target === overlay) closeModal(); });
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && !overlay.classList.contains('hidden')) closeModal();
    });

    // Делегирование на кнопки "Подробнее"
    document.addEventListener('click', (e) => {
        const btn = e.target.closest('.js-contact-details');
        if (!btn) return;

        const d = btn.dataset;

        currentContactId = d.id;

        document.querySelectorAll('.js-contact-details').forEach(el => el.classList.remove('active'));
        btn.classList.add('active');

        f.name.textContent = d.name || '—';

        f.phone.textContent = d.phone || '—';
        f.phone.href = d.phone ? ('tel:' + d.phone) : '#';

        if (d.email) {
        f.email.textContent = d.email;
        f.email.href = 'mailto:' + d.email;
        } else {
        f.email.textContent = '—';
        f.email.removeAttribute('href');
        }

        f.company.textContent = d.company || '—';
        f.type.textContent = d.type || '—';
        f.date.textContent = d.date || '—';

        const status = (d.status || 'new');
        f.status.textContent = (status === 'processed') ? 'Обработано' : 'Новая';
        f.status.className = 'badge ' + ((status === 'processed') ? 'badge-ok' : 'badge-blue');

        f.message.textContent = d.message || '—';

        openModal();
    });
    // === Универсальная смена статуса из модалки ===
    async function updateStatus(id, status) {
    const res = await fetch(`/admin/contacts/${id}/status`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status })
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.error || 'Не удалось изменить статус');
    return data;
    }

    function applyStatusToUI(id, status) {
    // 1) бейдж в модалке
    if (status === 'processed') {
        f.status.textContent = 'Обработано';
        f.status.className = 'badge badge-ok';
    } else if (status === 'new') {
        f.status.textContent = 'Новая';
        f.status.className = 'badge badge-blue';
    } else {
        f.status.textContent = 'В архиве';
        f.status.className = 'badge';
    }

    // 2) строка в таблице
    const tr = document.querySelector(`tr[data-id="${id}"]`);
    if (tr) {
        // обновим дата-статус у кнопки "Подробнее"
        const detailsBtn = tr.querySelector('.js-contact-details');
        if (detailsBtn) detailsBtn.dataset.status = status;

        // перерисуем ячейку статуса
        const cell = tr.querySelector('.status-cell');
        if (cell) {
        if (status === 'processed') {
            cell.innerHTML = '<span class="badge badge-ok">Обработано</span>';
        } else if (status === 'new') {
            cell.innerHTML = '<button class="btn btn-small mark-done" type="button">Обработать</button>';
        } else {
            cell.innerHTML = '<span class="badge">В архиве</span>';
        }
        }
    }
    }

    // кнопки в модалке
    const btnProcessed = document.getElementById('cd-mark-processed');
    const btnNew       = document.getElementById('cd-mark-new');
    const btnArchive   = document.getElementById('cd-archive');

    btnProcessed?.addEventListener('click', async () => {
    if (!currentContactId) return;
    try {
        await updateStatus(currentContactId, 'processed');
        applyStatusToUI(currentContactId, 'processed');
    } catch (e) {
        showAdminMessage?.(e.message, 'error');
    }
    });

    btnNew?.addEventListener('click', async () => {
    if (!currentContactId) return;
    try {
        await updateStatus(currentContactId, 'new');
        applyStatusToUI(currentContactId, 'new');
    } catch (e) {
        showAdminMessage?.(e.message, 'error');
    }
    });

    btnArchive?.addEventListener('click', async () => {
    if (!currentContactId) return;
    try {
        await updateStatus(currentContactId, 'archived');
        applyStatusToUI(currentContactId, 'archived');
    } catch (e) {
        showAdminMessage?.(e.message, 'error');
    }
    });

    // === Фильтры ===
    document.getElementById('apply-filters')?.addEventListener('click', applyFilters);
    document.getElementById('search-input')?.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            applyFilters();
        }
    });

    function applyFilters() {
        const search = document.getElementById('search-input')?.value || '';
        const status = document.getElementById('status-filter')?.value || '';
        const date   = document.getElementById('date-filter')?.value || '';

        const params = new URLSearchParams();
        if (search) params.set('search', search);
        if (status) params.set('status', status);
        if (date) params.set('date', date);

        window.location = '/admin/contacts' + (params.toString() ? '?' + params.toString() : '');
    }
});