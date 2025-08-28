// Массовые действия — АРХИВ
(function (w) {
    function initArchiveBulk() {
        // Общая привязка селекторов и кнопок
        w.ContactsShared.wireSelection(['#bulk-restore', '#bulk-delete']);

        document.getElementById('bulk-restore')?.addEventListener('click', async () => {
        const ids = w.ContactsShared.getSelectedIds(); if (!ids.length) return;
        await fetch('/admin/contacts/bulk', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ action: 'new', ids })
        }).then(r => r.json());
        ids.forEach(id => w.ContactsUI.removeRowById(id));
        w.ContactsShared.clearSelection();
        w.ContactsUI.show('ok', 'Восстановлено');
        });

        document.getElementById('bulk-delete')?.addEventListener('click', async () => {
        const ids = w.ContactsShared.getSelectedIds(); if (!ids.length) return;
        if (!confirm(`Удалить безвозвратно ${ids.length} шт.?`)) return;
        await Promise.all(ids.map(id => w.ContactsAPI.remove(id, { hard: true })));
        ids.forEach(id => w.ContactsUI.removeRowById(id));
        w.ContactsShared.clearSelection();
        w.ContactsUI.show('ok', 'Удалены');
        });
    }

    w.ContactsArchiveBulkInit = initArchiveBulk;
})(window);
