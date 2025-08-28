(function () {
    function applyFilters() {
        const search = document.getElementById('search-input')?.value || '';
        const date   = document.getElementById('date-filter')?.value || '';
        const limit  = document.getElementById('limit-select')?.value || '';

        const params = new URLSearchParams();
        if (search) params.set('search', search);
        if (date) params.set('date', date);
        if (limit) params.set('limit', limit);
        params.set('status', 'archived'); // фиксируем
        params.set('page', '1');

        window.location = '/admin/contacts/archive' + (params.toString() ? '?' + params.toString() : '');
    }

    document.addEventListener('DOMContentLoaded', () => {
        document.getElementById('apply-filters')?.addEventListener('click', applyFilters);
        document.getElementById('search-input')?.addEventListener('keydown', (e) => { if (e.key === 'Enter') applyFilters(); });

        // Экспорт CSV — как в обычной странице, backend уже учитывает archived
        document.getElementById('export-csv')?.addEventListener('click', () => {
        const url = window.ContactsAPI.exportUrlFromLocation();
        window.open(url, '_blank');
        });
    });
})();
