// Логика "Смотреть еще" для цен
document.addEventListener('DOMContentLoaded', function() {
  var grid = document.querySelector('.prices-grid');
  var loadMoreBtn = document.getElementById('pricesLoadMoreBtn');
  var loadMoreContainer = document.getElementById('pricesLoadMoreContainer');

  if (!grid || !loadMoreBtn || !loadMoreContainer) return;

  var allCards = Array.from(grid.querySelectorAll('.price-card'));
  var INITIAL_ROWS = 4;
  var LOAD_MORE_ROWS = 2;
  var currentlyShowing = getInitialItemsToShow();

  function getInitialItemsToShow() {
    return window.getGridColumns() * INITIAL_ROWS;
  }

  // Возвращает только карточки не скрытые фильтром
  function getVisibleCards() {
    return allCards.filter(function(card) {
      return !card.hasAttribute('data-filtered-out');
    });
  }

  function showPrices() {
    var visible = getVisibleCards();
    visible.forEach(function(card, index) {
      card.style.display = index < currentlyShowing ? '' : 'none';
    });

    loadMoreContainer.style.display = visible.length <= currentlyShowing ? 'none' : 'flex';

    if (window.centerGrid) window.centerGrid(grid, '.price-card');
  }

  loadMoreBtn.addEventListener('click', function() {
    currentlyShowing += window.getGridColumns() * LOAD_MORE_ROWS;
    showPrices();
  });

  // При смене фильтра — сбрасываем счётчик и пересчитываем
  window.addEventListener('pricesFiltered', function() {
    currentlyShowing = getInitialItemsToShow();
    showPrices();
  });

  var resizeTimer;
  window.addEventListener('resize', function() {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(function() {
      currentlyShowing = getInitialItemsToShow();
      showPrices();
    }, 250);
  });

  showPrices();
});
