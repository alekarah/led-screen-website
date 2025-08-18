document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('.mark-done').forEach(btn => {
        btn.addEventListener('click', async function () {
        const tr = this.closest('tr');
        const id = tr.dataset.id;

        try {
            const res = await fetch(`/admin/contacts/${id}/done`, { method: 'POST' });
            const data = await res.json();
            if (res.ok) {
            // заменяем кнопку на бейдж
            tr.querySelector('td:last-child').innerHTML =
                '<span class="badge badge-ok">Обработано</span>';
            if (window.showAdminMessage) showAdminMessage(data.message, 'success');
            } else {
            throw new Error(data.error || 'Ошибка');
            }
        } catch (e) {
            console.error(e);
            if (window.showAdminMessage) showAdminMessage(e.message, 'error');
        }
        });
    });
});
