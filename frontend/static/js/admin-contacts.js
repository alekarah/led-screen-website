document.addEventListener('DOMContentLoaded', () => {
  // ---------- Делегирование: "Обработать" ----------
  document.addEventListener('click', async (e) => {
    const btn = e.target.closest('.mark-done');
    if (!btn) return;

    const tr = btn.closest('tr');
    const id = tr?.dataset?.id;
    if (!id) return;

    btn.disabled = true;

    try {
      const res = await fetch(`/admin/contacts/${id}/done`, { method: 'POST' });
      let data = {};
      try { data = await res.json(); } catch (_) {}

      if (!res.ok) throw new Error(data.error || 'Не удалось пометить как обработано');

      const cell = tr.querySelector('.status-cell');
      if (cell) cell.innerHTML = '<span class="badge badge-ok">Обработано</span>';

      if (window.showAdminMessage) showAdminMessage(data.message || 'Заявка помечена как обработанная', 'success');
    } catch (err) {
      console.error(err);
      if (window.showAdminMessage) showAdminMessage(err.message, 'error');
    } finally {
      btn.disabled = false;
    }
  });

  // ---------- Модалка "Подробнее" ----------
  const overlay = document.getElementById('contact-details-modal');
  if (!overlay) return;

  const closeBtn = overlay.querySelector('.modal-close');
  const f = {
    name: document.getElementById('cd-name'),
    phone: document.getElementById('cd-phone'),
    email: document.getElementById('cd-email'),
    company: document.getElementById('cd-company'),
    type: document.getElementById('cd-type'),
    date: document.getElementById('cd-date'),
    status: document.getElementById('cd-status'),
    message: document.getElementById('cd-message'),
  };

  let lastFocused = null;

  function openModal() {
    lastFocused = document.activeElement;
    document.body.classList.add('modal-open');
    overlay.classList.remove('hidden');
    overlay.setAttribute('aria-hidden', 'false');
    closeBtn?.focus(); // удобнее закрыть сразу
  }

  function closeModal() {
    overlay.classList.add('hidden');
    overlay.setAttribute('aria-hidden', 'true');
    document.body.classList.remove('modal-open');
    lastFocused?.focus?.();
  }

  // закрытия
  if (closeBtn) closeBtn.addEventListener('click', closeModal);
  overlay.addEventListener('click', (e) => { if (e.target === overlay) closeModal(); });
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && !overlay.classList.contains('hidden')) closeModal();
  });

  // Делегирование на кнопки "Подробнее"
  document.addEventListener('click', (e) => {
    const btn = e.target.closest('.js-contact-details');
    if (!btn) return;

    const d = btn.dataset;

    f.name.textContent = d.name || '—';

    f.phone.textContent = d.phone || '—';
    f.phone.href = d.phone ? ('tel:' + d.phone) : '#';

    if (d.email) {
      f.email.textContent = d.email;
      f.email.href = 'mailto:' + d.email;
    } else {
      f.email.textContent = '—';
      f.email.removeAttribute('href');
    }

    f.company.textContent = d.company || '—';
    f.type.textContent = d.type || '—';
    f.date.textContent = d.date || '—';

    const processed = d.processed === 'true';
    f.status.textContent = processed ? 'Обработано' : 'Новая';
    f.status.className = 'badge ' + (processed ? 'badge-ok' : 'badge-blue');

    f.message.textContent = d.message || '—';

    openModal();
  });
});