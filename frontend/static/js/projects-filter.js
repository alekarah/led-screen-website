// Логика фильтрации проектов с toggle-поведением
document.addEventListener('DOMContentLoaded', function() {
    var filterBtns = document.querySelectorAll('.filters .filter-btn');
    var grid = document.getElementById('projectsGrid');

    if (!filterBtns.length || !grid) return;

    var cards = Array.from(grid.querySelectorAll('.public-project-card'));

    // Получаем категории каждой карточки из data-атрибутов или классов
    function getCardCategories(card) {
        var categoriesContainer = card.querySelector('.project-categories');
        if (!categoriesContainer) return [];

        var categoryTags = categoriesContainer.querySelectorAll('.category-tag');
        var categories = [];

        categoryTags.forEach(function(tag) {
            // Получаем slug из data-атрибута или текста
            var slug = tag.getAttribute('data-slug');
            if (slug) {
                categories.push(slug);
            }
        });

        return categories;
    }

    // Показать все проекты
    function showAllProjects() {
        cards.forEach(function(card) {
            card.style.display = '';
        });
        // Триггерим событие для пересчёта "Смотреть еще"
        window.dispatchEvent(new Event('projectsFiltered'));
    }

    // Фильтровать проекты по категории
    function filterByCategory(categorySlug) {
        cards.forEach(function(card) {
            var cardCategories = getCardCategories(card);
            if (cardCategories.indexOf(categorySlug) !== -1) {
                card.style.display = '';
            } else {
                card.style.display = 'none';
            }
        });
        // Триггерим событие для пересчёта "Смотреть еще"
        window.dispatchEvent(new Event('projectsFiltered'));
    }

    // Обработчик клика по фильтру
    filterBtns.forEach(function(btn) {
        btn.addEventListener('click', function(e) {
            e.preventDefault();

            var category = btn.getAttribute('data-category');
            var isActive = btn.classList.contains('active');

            // Убираем active со всех кнопок
            filterBtns.forEach(function(b) {
                b.classList.remove('active');
                b.removeAttribute('aria-current');
            });

            if (isActive) {
                // Повторный клик - снимаем фильтр, показываем все
                showAllProjects();
                // Обновляем URL без параметра категории
                history.pushState(null, '', '/projects');
            } else {
                // Клик по неактивному - активируем фильтр
                btn.classList.add('active');
                btn.setAttribute('aria-current', 'page');
                filterByCategory(category);
                // Обновляем URL с параметром категории
                history.pushState(null, '', '/projects?category=' + category);
            }
        });
    });

    // При загрузке страницы проверяем URL
    var urlParams = new URLSearchParams(window.location.search);
    var initialCategory = urlParams.get('category');

    if (initialCategory) {
        // Если есть категория в URL - фильтруем
        filterBtns.forEach(function(btn) {
            if (btn.getAttribute('data-category') === initialCategory) {
                btn.classList.add('active');
                btn.setAttribute('aria-current', 'page');
            }
        });
        filterByCategory(initialCategory);
    }
});
