// Логика "Смотреть еще" для цен
document.addEventListener('DOMContentLoaded', function() {
  var grid = document.querySelector('.prices-grid');
  var loadMoreBtn = document.getElementById('pricesLoadMoreBtn');
  var loadMoreContainer = document.getElementById('pricesLoadMoreContainer');

  if (!grid || !loadMoreBtn || !loadMoreContainer) return;

  var cards = Array.from(grid.querySelectorAll('.price-card'));
  var INITIAL_ROWS = 4;
  var LOAD_MORE_ROWS = 2;
  var currentlyShowing = getInitialItemsToShow();

  function getInitialItemsToShow() {
    return window.getGridColumns() * INITIAL_ROWS;
  }

  function showPrices() {
    cards.forEach(function(card, index) {
      card.style.display = index < currentlyShowing ? '' : 'none';
    });

    loadMoreContainer.style.display = cards.length <= currentlyShowing ? 'none' : 'flex';

    if (window.centerGrid) window.centerGrid(grid, '.price-card');
  }

  loadMoreBtn.addEventListener('click', function() {
    currentlyShowing += window.getGridColumns() * LOAD_MORE_ROWS;
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
