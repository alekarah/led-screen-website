// UI helpers: единые функции для обновления DOM
(function (w) {
    function setModalStatusBadge(el, status) {
        if (!el) return;
        if (status === 'processed') { el.textContent = 'Обработано'; el.className = 'badge badge-ok'; }
        else if (status === 'new')  { el.textContent = 'Новая';      el.className = 'badge badge-blue'; }
        else                        { el.textContent = 'В архиве';   el.className = 'badge'; }
    }

    function setRowStatus(tr, status) {
        if (!tr) return;
        const detailsBtn = tr.querySelector('.js-contact-details');
        if (detailsBtn) detailsBtn.dataset.status = status;

        const cell = tr.querySelector('.status-cell');
        if (!cell) return;

        if (status === 'processed') {
        cell.innerHTML = '<span class="badge badge-ok">Обработано</span>';
        } else if (status === 'new') {
        cell.innerHTML = '<button class="btn btn-small mark-done" type="button">Обработать</button>';
        } else {
        cell.innerHTML = '<span class="badge">В архиве</span>';
        }
    }

    function setRowStatusById(id, status) {
        const tr = document.querySelector(`tr[data-id="${id}"]`);
        setRowStatus(tr, status);
    }

    function show(type, msg) {
        // type: 'ok' | 'error'
        window.showAdminMessage?.(msg, type === 'error' ? 'error' : 'ok');
    }

    w.ContactsUI = { setModalStatusBadge, setRowStatus, setRowStatusById, show };
})(window);
