// Единая точка входа: инициализируем модули в нужном порядке
(function () {
  document.addEventListener('DOMContentLoaded', () => {
    // порядок важен только логически — модули друг от друга не зависят,
    // но так читается и поддерживается лучше
    window.ContactsFiltersInit?.();
    window.ContactsModalInit?.();
    window.ContactsBulkInit?.();
  });
})();
