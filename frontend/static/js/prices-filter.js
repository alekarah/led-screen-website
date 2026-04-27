// Логика фильтрации позиций прайса с toggle-поведением
document.addEventListener('DOMContentLoaded', function() {
    var categoryBtns = document.querySelectorAll('.prices-filter-btn[data-category]');
    var lightBtn = document.querySelector('.prices-filter-btn[data-filter="light"]');
    var grid = document.getElementById('pricesGrid');

    if (!grid) return;

    var activeCategory = null;
    var activeLight = false;

    function applyFilters() {
        var cards = Array.from(grid.querySelectorAll('.price-card'));

        cards.forEach(function(card) {
            var catMatch = !activeCategory || card.getAttribute('data-category') === activeCategory;
            var lightMatch = !activeLight || card.getAttribute('data-light') === 'true';
            // data-filtered-out используется load-more для подсчёта видимых
            if (catMatch && lightMatch) {
                card.removeAttribute('data-filtered-out');
            } else {
                card.setAttribute('data-filtered-out', '1');
                card.style.display = 'none';
            }
        });

        // Пересчитываем load-more
        window.dispatchEvent(new Event('pricesFiltered'));
    }

    // Фильтры по категории — toggle
    categoryBtns.forEach(function(btn) {
        btn.addEventListener('click', function() {
            var cat = btn.getAttribute('data-category');
            if (activeCategory === cat) {
                activeCategory = null;
                btn.classList.remove('active');
            } else {
                categoryBtns.forEach(function(b) { b.classList.remove('active'); });
                activeCategory = cat;
                btn.classList.add('active');
            }
            applyFilters();
        });
    });

    // Фильтр Light — независимый toggle
    if (lightBtn) {
        lightBtn.addEventListener('click', function() {
            activeLight = !activeLight;
            lightBtn.classList.toggle('active', activeLight);
            applyFilters();
        });
    }
});
