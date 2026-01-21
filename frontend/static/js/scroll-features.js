// Scroll Features - Progress Bar и Back to Top для всех страниц

(function() {
    'use strict';

    // ===== Reading Progress Bar =====
    function initProgressBar() {
        var progressBar = document.querySelector('.reading-progress-bar');
        if (!progressBar) return;

        function updateProgressBar() {
            var windowHeight = window.innerHeight;
            var documentHeight = document.documentElement.scrollHeight;
            var scrollTop = window.pageYOffset || document.documentElement.scrollTop;

            // Если страница короткая - не показываем прогресс
            if (documentHeight <= windowHeight) {
                progressBar.style.width = '0%';
                return;
            }

            var scrollPercent = (scrollTop / (documentHeight - windowHeight)) * 100;
            progressBar.style.width = Math.min(scrollPercent, 100) + '%';
        }

        window.addEventListener('scroll', updateProgressBar, { passive: true });
        window.addEventListener('resize', updateProgressBar, { passive: true });
        updateProgressBar();
    }

    // ===== Back to Top Button =====
    function initBackToTop() {
        var backToTopBtn = document.getElementById('backToTop');
        if (!backToTopBtn) return;

        function toggleBackToTop() {
            if (window.pageYOffset > 300) {
                backToTopBtn.classList.add('visible');
            } else {
                backToTopBtn.classList.remove('visible');
            }
        }

        window.addEventListener('scroll', toggleBackToTop, { passive: true });

        backToTopBtn.addEventListener('click', function(e) {
            e.preventDefault();
            window.scrollTo({
                top: 0,
                behavior: 'smooth'
            });
        });

        toggleBackToTop();
    }

    // ===== Initialize =====
    function init() {
        initProgressBar();
        initBackToTop();
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
