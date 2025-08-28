// Массовые действия — ВХОДЯЩИЕ
(function (w) {
    function initBulk() {
        // Общая привязка селекторов и кнопок
        w.ContactsShared.wireSelection(['#bulk-process', '#bulk-restore', '#bulk-archive']);

        // Действия
        document.getElementById('bulk-process')?.addEventListener('click', async () => {
        const ids = w.ContactsShared.getSelectedIds(); if (!ids.length) return;
        try {
            await w.ContactsAPI.bulk('processed', ids);
            ids.forEach(id => w.ContactsUI.setRowStatusById(id, 'processed'));
            w.ContactsShared.clearSelection();
            w.ContactsUI.show('ok', 'Помечено как обработано');
        } catch (err) { w.ContactsUI.show('error', err.message); }
        });

        document.getElementById('bulk-restore')?.addEventListener('click', async () => {
        const ids = w.ContactsShared.getSelectedIds(); if (!ids.length) return;
        try {
            await w.ContactsAPI.bulk('new', ids);
            ids.forEach(id => w.ContactsUI.setRowStatusById(id, 'new'));
            w.ContactsShared.clearSelection();
            w.ContactsUI.show('ok', 'Возвращены в новые');
        } catch (err) { w.ContactsUI.show('error', err.message); }
        });

        document.getElementById('bulk-archive')?.addEventListener('click', async () => {
        const ids = w.ContactsShared.getSelectedIds(); if (!ids.length) return;
        try {
            await w.ContactsAPI.bulk('archived', ids);
            ids.forEach(id => w.ContactsUI.setRowStatusById(id, 'archived'));
            w.ContactsShared.clearSelection();
            w.ContactsUI.show('ok', 'Отправлено в архив');
        } catch (err) { w.ContactsUI.show('error', err.message); }
        });
    }

    w.ContactsBulkInit = initBulk;
})(window);
