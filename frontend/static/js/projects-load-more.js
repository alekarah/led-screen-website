// Логика "Смотреть еще" для проектов
document.addEventListener('DOMContentLoaded', function() {
  const grid = document.getElementById('projectsGrid');
  const loadMoreBtn = document.getElementById('loadMoreBtn');
  const loadMoreContainer = document.getElementById('loadMoreContainer');

  if (!grid || !loadMoreBtn || !loadMoreContainer) {
    return;
  }

  const cards = Array.from(grid.querySelectorAll('.public-project-card'));
  let currentlyShowing = getInitialItemsToShow();

  // Функция определения начального количества проектов
  function getInitialItemsToShow() {
    const width = window.innerWidth;
    if (width > 1200) return 12; // 4 ряда по 3
    if (width > 768) return 8;   // 4 ряда по 2
    return 4;                     // 4 ряда по 1
  }

  // Получить количество колонок в текущем брейкпоинте
  function getColumnsCount() {
    const width = window.innerWidth;
    if (width > 1200) return 3;
    if (width > 768) return 2;
    return 1;
  }

  // Показать проекты
  function showProjects() {
    // Сначала убираем все классы центрирования
    cards.forEach(card => {
      card.style.display = '';
      card.classList.remove('js-center-single-3col', 'js-center-double-3col-first', 'js-center-double-3col-second', 'js-center-single-2col');
    });

    // Показываем/скрываем карточки
    const visibleCards = [];
    cards.forEach((card, index) => {
      if (index < currentlyShowing) {
        card.style.display = '';
        visibleCards.push(card);
      } else {
        card.style.display = 'none';
      }
    });

    // Центрируем неполный последний ряд
    const columns = getColumnsCount();
    const visibleCount = visibleCards.length;
    const lastRowCount = visibleCount % columns;

    if (lastRowCount > 0 && columns > 1) {
      const firstInLastRow = visibleCount - lastRowCount;

      if (columns === 3) {
        // 3 колонки
        if (lastRowCount === 1) {
          // Одна карточка - во вторую колонку
          visibleCards[firstInLastRow].classList.add('js-center-single-3col');
        } else if (lastRowCount === 2) {
          // Две карточки - центрируем
          visibleCards[firstInLastRow].classList.add('js-center-double-3col-first');
          visibleCards[firstInLastRow + 1].classList.add('js-center-double-3col-second');
        }
      } else if (columns === 2 && lastRowCount === 1) {
        // 2 колонки, одна карточка - span на обе колонки и центр
        visibleCards[firstInLastRow].classList.add('js-center-single-2col');
      }
    }

    // Показать/скрыть кнопку
    if (cards.length <= currentlyShowing) {
      loadMoreContainer.style.display = 'none';
    } else {
      loadMoreContainer.style.display = 'flex';
    }
  }

  // Обработчик кнопки - показываем еще 4 проекта
  loadMoreBtn.addEventListener('click', function() {
    currentlyShowing += 4;
    showProjects();
  });

  // Обработчик ресайза
  let resizeTimer;
  window.addEventListener('resize', function() {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(function() {
      const newInitialCount = getInitialItemsToShow();
      // Сбрасываем к начальному значению при изменении размера окна
      currentlyShowing = newInitialCount;
      showProjects();
    }, 250);
  });

  // Инициализация
  showProjects();
});
