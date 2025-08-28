// API: все запросы к бэкенду в одном месте
(function (w) {

    async function request(url, { method = 'GET', body, headers } = {}) {
        const opts = { method, headers: headers || {} };
        if (body !== undefined) {
        opts.headers['Content-Type'] = 'application/json';
        opts.body = JSON.stringify(body);
        }
        const res = await fetch(url, opts);
        let data = {};
        try { data = await res.json(); } catch (_) {}
        if (!res.ok || data.success === false) {
        const msg = data.error || data.message || `HTTP ${res.status}`;
        throw new Error(msg);
        }
        return data;
    }

    const api = {
        request, // пусть будет доступен, вдруг пригодится где-то ещё

        updateStatus(id, status) {
        return request(`/admin/contacts/${id}/status`, { method: 'POST', body: { status } });
        },

        bulk(action, ids) {
        return request('/admin/contacts/bulk', { method: 'POST', body: { action, ids } });
        },

        archive(id) {
        return request(`/admin/contacts/${id}/archive`, { method: 'PATCH' });
        },

        restore(id, to = 'new') {
        return request(`/admin/contacts/${id}/restore`, { method: 'PATCH', body: { to } });
        },

        remove(id, { hard = true } = {}) {
        const url = `/admin/contacts/${id}` + (hard ? '?hard=true' : '');
        return request(url, { method: 'DELETE' });
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
