// Единая точка входа
(function () {
  document.addEventListener('DOMContentLoaded', () => {
    window.ContactsFiltersInit?.();
    window.ContactsModalInit?.();
    window.ContactsBulkInit?.();

    // Делегирование: "Обработать" в строке
    document.querySelector('table.table tbody')?.addEventListener('click', async (e) => {
      const btn = e.target.closest('.mark-done');
      if (!btn) return;
      const tr = btn.closest('tr');
      const id = Number(tr?.dataset?.id || 0);
      if (!id) return;
      try {
        await window.ContactsAPI.updateStatus(id, 'processed');
        window.ContactsUI.setRowStatus(tr, 'processed');
        window.ContactsUI.show('ok', 'Помечено как обработано');
      } catch (err) {
        window.ContactsUI.show('error', err.message);
      }
    });
  });
})();
