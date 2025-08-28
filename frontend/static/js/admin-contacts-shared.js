// Общие утилиты выбора строк и управления bulk-кнопками
(function (w) {
    const $  = (s) => document.querySelector(s);
    const $$ = (s) => Array.from(document.querySelectorAll(s));

    let recalc = () => {};

    function getSelectedIds() {
        return $$('.row-select:checked').map(ch => Number(ch.value));
    }

    function updateBulkCount(n) {
        const el = $('#bulk-count');
        if (el) el.textContent = `Выбрано: ${n}`;
    }

    function setDisabled(selectors, disabled) {
        selectors.forEach((id) => { const el = $(id); if (el) el.disabled = disabled; });
    }

    // Привязывает select-all и чекбоксы строк; включает/выключает кнопки
    function wireSelection(buttonSelectors) {
        recalc = () => {
        const n = getSelectedIds().length;
        updateBulkCount(n);
        setDisabled(buttonSelectors, n === 0);
        };

        $('#select-all')?.addEventListener('change', (e) => {
        const checked = e.target.checked;
        $$('.row-select').forEach(ch => ch.checked = checked);
        recalc();
        });

        $('table.table tbody')?.addEventListener('change', (e) => {
        if (!e.target.closest('.row-select')) return;
        if (!e.target.checked) { const sa = $('#select-all'); if (sa) sa.checked = false; }
        recalc();
        });

        recalc(); // первичная инициализация
    }

    function clearSelection() {
        const sa = $('#select-all');
        if (sa) sa.checked = false;
        $$('.row-select:checked').forEach(ch => ch.checked = false);
        recalc();
    }

    w.ContactsShared = { getSelectedIds, wireSelection, clearSelection };
})(window);
