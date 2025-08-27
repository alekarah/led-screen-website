// Фильтры и экспорт
(function (w) {
    function applyFilters() {
        const search = document.getElementById('search-input')?.value || '';
        const status = document.getElementById('status-filter')?.value || '';
        const date   = document.getElementById('date-filter')?.value || '';
        const limitSel = document.getElementById('limit-select');

        const params = new URLSearchParams();
        if (search) params.set('search', search);
        if (status) params.set('status', status);
        if (date)   params.set('date', date);
        if (limitSel && limitSel.value) params.set('limit', limitSel.value);
        params.set('page', '1');

        window.location = '/admin/contacts' + (params.toString() ? '?' + params.toString() : '');
    }

    function initFilters() {
        document.getElementById('apply-filters')?.addEventListener('click', applyFilters);

        // Enter в поле поиска
        document.getElementById('search-input')?.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') applyFilters();
        });

        // Экспорт CSV по текущим параметрам из URL
        document.getElementById('export-csv')?.addEventListener('click', () => {
        const url = window.ContactsAPI.exportUrlFromLocation();
        window.open(url, '_blank');
        });
    }

    w.ContactsFiltersInit = initFilters;
})(window);
