// API: все запросы к бэкенду в одном месте
(function (w) {
    const json = (res) => res.json().catch(() => ({}));

    const api = {
        async updateStatus(id, status) {
        const res = await fetch(`/admin/contacts/${id}/status`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status })
        });
        const data = await json(res);
        if (!res.ok) throw new Error(data.error || 'Не удалось изменить статус');
        return data;
        },

        async bulk(action, ids) {
        const res = await fetch('/admin/contacts/bulk', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ action, ids })
        });
        const data = await json(res);
        if (!res.ok || data.success === false) {
            throw new Error(data.error || 'Не удалось выполнить массовое действие');
        }
        return data;
        },

        exportUrlFromLocation() {
        const current = new URLSearchParams(location.search);
        const p = new URLSearchParams();
        if (current.get('search')) p.set('q', current.get('search'));
        else if (current.get('q')) p.set('q', current.get('q'));
        if (current.get('status')) p.set('status', current.get('status'));
        if (current.get('date'))   p.set('date',   current.get('date'));
        return '/admin/contacts/export.csv' + (p.toString() ? ('?' + p.toString()) : '');
        }
    };

    w.ContactsAPI = api;
})(window);
