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

        const params = new URLSearchParams();            // <-- сперва создаём

        if (search) params.set('search', search);
        if (status) params.set('status', status);
        if (date)   params.set('date', date);

        const limitSel = document.getElementById('limit-select');
        if (limitSel && limitSel.value) params.set('limit', limitSel.value);

        params.set('page', '1'); // при применении фильтров начать с первой страницы

        window.location = '/admin/contacts' + (params.toString() ? '?' + params.toString() : '');
    }

    // === Экспорт CSV (с текущими фильтрами из URL) ===
    document.getElementById('export-csv')?.addEventListener('click', () => {
    const current = new URLSearchParams(location.search);
    const p = new URLSearchParams();

    // поддерживаем и search, и q (на будущее)
    if (current.get('search')) p.set('q', current.get('search'));
    else if (current.get('q')) p.set('q', current.get('q'));

    if (current.get('status')) p.set('status', current.get('status'));
    if (current.get('date'))   p.set('date',   current.get('date'));

    const url = '/admin/contacts/export.csv' + (p.toString() ? ('?' + p.toString()) : '');
    window.open(url, '_blank'); // скачивание в новой вкладке
    });

    // === Массовые действия: выбор чекбоксов ===
    const selectAll = document.getElementById('select-all');
    const bulkCount = document.getElementById('bulk-count');
    const btnBulkProcess = document.getElementById('bulk-process');
    const btnBulkArchive = document.getElementById('bulk-archive');
    const btnBulkRestore = document.getElementById('bulk-restore');

    // соберём id из отмеченных строк
    function getSelectedIds() {
        const ids = [];
        document.querySelectorAll('input.row-select:checked').forEach(ch => {
            const val = ch.value?.trim();
            if (val) ids.push(Number(val));
        });
        return ids;
    }

    function updateBulkUI() {
        const ids = getSelectedIds();
        const n = ids.length;
        if (bulkCount) bulkCount.textContent = `Выбрано: ${n}`;
        const enabled = n > 0;
        if (btnBulkProcess) btnBulkProcess.disabled = !enabled;
        if (btnBulkArchive) btnBulkArchive.disabled = !enabled;
        if (btnBulkRestore) btnBulkRestore.disabled = !enabled;
    }

    // «выбрать все на странице»
    selectAll?.addEventListener('change', () => {
        const checked = !!selectAll.checked;
        document.querySelectorAll('input.row-select').forEach(ch => ch.checked = checked);
        updateBulkUI();
    });

    // переключение отдельных строк (делегирование на tbody — так быстрее)
    document.querySelector('table.table tbody')?.addEventListener('change', (e) => {
        const ch = e.target.closest('input.row-select');
        if (!ch) return;
        // если сняли галочку у строки — снимаем и «выбрать все»
        if (!ch.checked && selectAll && selectAll.checked) {
            selectAll.checked = false;
        }
        updateBulkUI();
    });

    // первичная инициализация
    updateBulkUI();

    // унифицированный вызов bulk-операции
    async function bulkAction(action, ids) {
        const res = await fetch('/admin/contacts/bulk', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ action, ids })
        });
        const data = await res.json().catch(() => ({}));
        if (!res.ok || data.success === false) {
            throw new Error(data.error || 'Не удалось выполнить массовое действие');
        }
        return data;
    }

    // обновление UI строки статуса (повторяем логику, как при одиночном действии)
    function updateRowStatusUI(id, status) {
        const tr = document.querySelector(`tr[data-id="${id}"]`);
        if (!tr) return;

        // кнопка «Подробнее» хранит статус в data-status — обновим
        const detailsBtn = tr.querySelector('.js-contact-details');
        if (detailsBtn) detailsBtn.dataset.status = status;

        // текст/кнопка в ячейке статуса
        const cell = tr.querySelector('.status-cell');
        if (!cell) return;

        if (status === 'processed') {
            cell.innerHTML = '<span class="badge badge-ok">Обработано</span>';
        } else if (status === 'new') {
            cell.innerHTML = '<button class="btn btn-small mark-done" type="button">Обработать</button>';
        } else { // archived
            cell.innerHTML = '<span class="badge">В архиве</span>';
        }
    }

    // навешиваем обработчики на панель
    btnBulkProcess?.addEventListener('click', async () => {
        const ids = getSelectedIds();
        if (!ids.length) return;
        try {
            btnBulkProcess.disabled = true;
            await bulkAction('processed', ids);
            ids.forEach(id => updateRowStatusUI(id, 'processed'));
            // снимаем выделение
            document.querySelectorAll('input.row-select:checked').forEach(ch => ch.checked = false);
            if (selectAll) selectAll.checked = false;
            updateBulkUI();
            showAdminMessage?.('Помечено как обработано', 'ok');
        } catch (e) {
            showAdminMessage?.(e.message, 'error');
        } finally {
            updateBulkUI();
        }
    });

    btnBulkRestore?.addEventListener('click', async () => {
        const ids = getSelectedIds();
        if (!ids.length) return;
        try {
            btnBulkRestore.disabled = true;
            await bulkAction('new', ids);
            ids.forEach(id => updateRowStatusUI(id, 'new'));
            document.querySelectorAll('input.row-select:checked').forEach(ch => ch.checked = false);
            if (selectAll) selectAll.checked = false;
            updateBulkUI();
            showAdminMessage?.('Возвращены в новые', 'ok');
        } catch (e) {
            showAdminMessage?.(e.message, 'error');
        } finally {
            updateBulkUI();
        }
    });

    btnBulkArchive?.addEventListener('click', async () => {
        const ids = getSelectedIds();
        if (!ids.length) return;
        try {
            btnBulkArchive.disabled = true;
            await bulkAction('archived', ids);
            ids.forEach(id => updateRowStatusUI(id, 'archived'));
            document.querySelectorAll('input.row-select:checked').forEach(ch => ch.checked = false);
            if (selectAll) selectAll.checked = false;
            updateBulkUI();
            showAdminMessage?.('Отправлено в архив', 'ok');
        } catch (e) {
            showAdminMessage?.(e.message, 'error');
        } finally {
            updateBulkUI();
        }
    });
});