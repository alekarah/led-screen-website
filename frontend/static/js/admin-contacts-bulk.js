// Чекбоксы и массовые действия
(function (w) {
    const $ = (sel) => document.querySelector(sel);
    const $$ = (sel) => Array.from(document.querySelectorAll(sel));

    function getSelectedIds() {
        return $$('.row-select:checked').map(ch => Number(ch.value));
    }

    function updateBulkUI() {
        const n = getSelectedIds().length;
        $('#bulk-count') && ($('#bulk-count').textContent = `Выбрано: ${n}`);
        ['#bulk-process','#bulk-restore','#bulk-archive'].forEach(id => {
        const el = $(id);
        if (el) el.disabled = n === 0;
        });
    }

    function clearSelection() {
        $('#select-all') && ($('#select-all').checked = false);
        $$('.row-select:checked').forEach(ch => ch.checked = false);
        updateBulkUI();
    }

    function initBulk() {
        // select all
        $('#select-all')?.addEventListener('change', (e) => {
        const checked = e.target.checked;
        $$('.row-select').forEach(ch => ch.checked = checked);
        updateBulkUI();
        });

        // выбор отдельных строк (делегирование)
        $('table.table tbody')?.addEventListener('change', (e) => {
        if (!e.target.closest('.row-select')) return;
        if (!e.target.checked) { const sa = $('#select-all'); if (sa) sa.checked = false; }
        updateBulkUI();
        });

        // кнопка «Обработать»
        $('#bulk-process')?.addEventListener('click', async () => {
        const ids = getSelectedIds(); if (!ids.length) return;
        try {
            await window.ContactsAPI.bulk('processed', ids);
            ids.forEach(id => window.ContactsUI.setRowStatusById(id, 'processed'));
            clearSelection();
            window.ContactsUI.show('ok','Помечено как обработано');
        } catch (err) { window.ContactsUI.show('error', err.message); }
        });

        // «Вернуть в новые»
        $('#bulk-restore')?.addEventListener('click', async () => {
        const ids = getSelectedIds(); if (!ids.length) return;
        try {
            await window.ContactsAPI.bulk('new', ids);
            ids.forEach(id => window.ContactsUI.setRowStatusById(id, 'new'));
            clearSelection();
            window.ContactsUI.show('ok','Возвращены в новые');
        } catch (err) { window.ContactsUI.show('error', err.message); }
        });

        // «Архивировать»
        $('#bulk-archive')?.addEventListener('click', async () => {
        const ids = getSelectedIds(); if (!ids.length) return;
        try {
            await window.ContactsAPI.bulk('archived', ids);
            ids.forEach(id => window.ContactsUI.setRowStatusById(id, 'archived'));
            clearSelection();
            window.ContactsUI.show('ok','Отправлено в архив');
        } catch (err) { window.ContactsUI.show('error', err.message); }
        });

        updateBulkUI(); // первичная инициализация
    }

    w.ContactsBulkInit = initBulk;
})(window);
