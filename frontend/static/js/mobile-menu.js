// Управление мобильным меню
(function() {
    const burgerBtn = document.querySelector('.burger-menu');
    const navLinks = document.querySelector('.nav-links');
    const body = document.body;

    if (!burgerBtn || !navLinks) return;

    // Открытие/закрытие меню по клику на бургер
    burgerBtn.addEventListener('click', function() {
        const isOpen = navLinks.classList.contains('active');

        if (isOpen) {
            closeMenu();
        } else {
            openMenu();
        }
    });

    // Закрытие меню при клике на ссылку
    const menuLinks = navLinks.querySelectorAll('a');
    menuLinks.forEach(link => {
        link.addEventListener('click', function() {
            closeMenu();
        });
    });

    // Закрытие меню при клике на оверлей
    body.addEventListener('click', function(e) {
        if (body.classList.contains('menu-open') &&
            !navLinks.contains(e.target) &&
            !burgerBtn.contains(e.target)) {
            closeMenu();
        }
    });

    // Закрытие меню по Escape
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape' && body.classList.contains('menu-open')) {
            closeMenu();
        }
    });

    function openMenu() {
        navLinks.classList.add('active');
        burgerBtn.classList.add('active');
        body.classList.add('menu-open');
        burgerBtn.setAttribute('aria-expanded', 'true');
        burgerBtn.setAttribute('aria-label', 'Закрыть меню');
    }

    function closeMenu() {
        navLinks.classList.remove('active');
        burgerBtn.classList.remove('active');
        body.classList.remove('menu-open');
        burgerBtn.setAttribute('aria-expanded', 'false');
        burgerBtn.setAttribute('aria-label', 'Открыть меню');
    }
})();
