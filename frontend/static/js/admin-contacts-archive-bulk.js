(function (w) {
  const $ = (s) => document.querySelector(s);
  const $$ = (s) => Array.from(document.querySelectorAll(s));

  function getSelectedIds() {
    return $$('.row-select:checked').map(ch => Number(ch.value));
  }

  function updateBulkUI() {
    const n = getSelectedIds().length;
    $('#bulk-count') && ($('#bulk-count').textContent = `Выбрано: ${n}`);
    ['#bulk-restore','#bulk-delete'].forEach(id => { const el = $(id); if (el) el.disabled = n === 0; });
  }

  function clearSelection() {
    $('#select-all') && ($('#select-all').checked = false);
    $$('.row-select:checked').forEach(ch => ch.checked = false);
    updateBulkUI();
  }

  async function bulkDelete(ids) {
    if (!ids.length) return;
    if (!confirm(`Удалить безвозвратно ${ids.length} шт.?`)) return;
    await Promise.all(ids.map(id => w.ContactsAPI.remove(id, { hard: true })));
    ids.forEach(id => w.ContactsUI.removeRowById(id));
    clearSelection();
    w.ContactsUI.show('ok', 'Удалены');
  }

  async function bulkRestore(ids) {
    if (!ids.length) return;
    await fetch('/admin/contacts/bulk', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ action: 'new', ids })
    }).then(r => r.json());
    ids.forEach(id => w.ContactsUI.removeRowById(id));
    clearSelection();
    w.ContactsUI.show('ok', 'Восстановлено');
  }

  function initArchiveBulk() {
    // select all
    $('#select-all')?.addEventListener('change', (e) => {
      const checked = e.target.checked;
      $$('.row-select').forEach(ch => ch.checked = checked);
      updateBulkUI();
    });

    // выбор отдельных строк
    $('table.table tbody')?.addEventListener('change', (e) => {
      if (!e.target.closest('.row-select')) return;
      if (!e.target.checked) { const sa = $('#select-all'); if (sa) sa.checked = false; }
      updateBulkUI();
    });

    $('#bulk-restore')?.addEventListener('click', () => bulkRestore(getSelectedIds()));
    $('#bulk-delete')?.addEventListener('click', () => bulkDelete(getSelectedIds()));
    updateBulkUI();
  }

  w.ContactsArchiveBulkInit = initArchiveBulk;
})(window);
