// Универсальное центрирование неполного последнего ряда в гриде
(function() {
  var CENTER_CLASSES = ['js-center-single-3col', 'js-center-double-3col-first', 'js-center-double-3col-second', 'js-center-single-2col'];

  function getColumnsCount() {
    var w = window.innerWidth;
    if (w > 1200) return 3;
    if (w > 768) return 2;
    return 1;
  }

  // Центрирует видимые карточки в гриде
  // grid — элемент-контейнер, cardSelector — селектор карточек
  function centerGrid(grid, cardSelector) {
    if (!grid) return;
    var cards = Array.from(grid.querySelectorAll(cardSelector));

    // Собираем только видимые карточки
    var visible = [];
    cards.forEach(function(c) {
      CENTER_CLASSES.forEach(function(cls) { c.classList.remove(cls); });
      if (c.style.display !== 'none') visible.push(c);
    });

    var columns = getColumnsCount();
    var count = visible.length;
    var lastRowCount = count % columns;

    if (lastRowCount === 0 || columns <= 1) return;

    var first = count - lastRowCount;

    if (columns === 3) {
      if (lastRowCount === 1) {
        visible[first].classList.add('js-center-single-3col');
      } else if (lastRowCount === 2) {
        visible[first].classList.add('js-center-double-3col-first');
        visible[first + 1].classList.add('js-center-double-3col-second');
      }
    } else if (columns === 2 && lastRowCount === 1) {
      visible[first].classList.add('js-center-single-2col');
    }
  }

  // Экспортируем глобально
  window.centerGrid = centerGrid;
  window.getGridColumns = getColumnsCount;
})();
