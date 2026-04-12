// Логика "Смотреть еще" для проектов
document.addEventListener('DOMContentLoaded', function() {
  var grid = document.getElementById('projectsGrid');
  var loadMoreBtn = document.getElementById('loadMoreBtn');
  var loadMoreContainer = document.getElementById('loadMoreContainer');

  if (!grid || !loadMoreBtn || !loadMoreContainer) return;

  var cards = Array.from(grid.querySelectorAll('.public-project-card'));
  var INITIAL_ROWS = 4;
  var LOAD_MORE_ROWS = 2;
  var currentlyShowing = getInitialItemsToShow();

  function getInitialItemsToShow() {
    return window.getGridColumns() * INITIAL_ROWS;
  }

  // Получаем карточки, подходящие под текущий фильтр
  function getFilteredCards() {
    var activeFilter = document.querySelector('.filters .filter-btn.active');
    if (!activeFilter) return cards;

    var category = activeFilter.getAttribute('data-category');
    return cards.filter(function(card) {
      var tagsContainer = card.querySelector('.project-categories');
      if (!tagsContainer) return false;
      var tags = tagsContainer.querySelectorAll('.category-tag');
      for (var i = 0; i < tags.length; i++) {
        if (tags[i].getAttribute('data-slug') === category) return true;
      }
      return false;
    });
  }

  function showProjects() {
    var filtered = getFilteredCards();

    // Если в URL есть hash — находим нужную карточку среди всех карточек
    var hashSlug = window.location.hash ? decodeURIComponent(window.location.hash.substring(1)) : null;
    var hashCardEl = null;
    if (hashSlug) {
      for (var i = 0; i < cards.length; i++) {
        var btn = cards[i].querySelector('[data-project-slug="' + hashSlug + '"]');
        if (btn) { hashCardEl = cards[i]; break; }
      }
    }

    // Сначала скрываем все
    cards.forEach(function(card) { card.style.display = 'none'; });

    // Показываем отфильтрованные до лимита
    filtered.forEach(function(card, index) {
      card.style.display = index < currentlyShowing ? '' : 'none';
    });

    // Принудительно показываем карточку из hash
    if (hashCardEl) { hashCardEl.style.display = ''; }

    loadMoreContainer.style.display = filtered.length <= currentlyShowing ? 'none' : 'flex';

    if (window.centerGrid) window.centerGrid(grid, '.public-project-card');
  }

  loadMoreBtn.addEventListener('click', function() {
    currentlyShowing += window.getGridColumns() * LOAD_MORE_ROWS;
    showProjects();
  });

  // Пересчёт при фильтрации
  window.addEventListener('projectsFiltered', function() {
    currentlyShowing = getInitialItemsToShow();
    showProjects();
  });

  var resizeTimer;
  window.addEventListener('resize', function() {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(function() {
      currentlyShowing = getInitialItemsToShow();
      showProjects();
    }, 250);
  });

  showProjects();
});
