// Логика фильтрации проектов с toggle-поведением
document.addEventListener('DOMContentLoaded', function() {
    var filterBtns = document.querySelectorAll('.filters .filter-btn');
    var grid = document.getElementById('projectsGrid');

    if (!filterBtns.length || !grid) return;

    // Обработчик клика по фильтру
    filterBtns.forEach(function(btn) {
        btn.addEventListener('click', function(e) {
            e.preventDefault();

            var isActive = btn.classList.contains('active');

            // Убираем active со всех кнопок
            filterBtns.forEach(function(b) {
                b.classList.remove('active');
                b.removeAttribute('aria-current');
            });

            if (isActive) {
                // Повторный клик — снимаем фильтр
                history.pushState(null, '', '/projects');
            } else {
                // Активируем фильтр
                btn.classList.add('active');
                btn.setAttribute('aria-current', 'page');
                var category = btn.getAttribute('data-category');
                history.pushState(null, '', '/projects?category=' + category);
            }

            // Load-more пересчитает видимость
            window.dispatchEvent(new Event('projectsFiltered'));
        });
    });

    // При загрузке страницы проверяем URL
    var urlParams = new URLSearchParams(window.location.search);
    var initialCategory = urlParams.get('category');

    if (initialCategory) {
        filterBtns.forEach(function(btn) {
            if (btn.getAttribute('data-category') === initialCategory) {
                btn.classList.add('active');
                btn.setAttribute('aria-current', 'page');
            }
        });
        // Load-more при инициализации сам прочитает активный фильтр
    }
});
