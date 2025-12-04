// FAQ Accordion функциональность
(function() {
    'use strict';

    function initFAQ() {
        const faqItems = document.querySelectorAll('.faq-item');

        if (!faqItems || faqItems.length === 0) {
            return;
        }

        faqItems.forEach(function(item) {
            const question = item.querySelector('.faq-question');

            if (!question) {
                return;
            }

            // Добавляем обработчик клика
            question.addEventListener('click', function(e) {
                e.preventDefault();

                var isCurrentlyActive = item.classList.contains('active');

                // Закрываем все элементы
                faqItems.forEach(function(otherItem) {
                    otherItem.classList.remove('active');
                });

                // Если элемент был закрыт, открываем его
                if (!isCurrentlyActive) {
                    item.classList.add('active');
                }
            });

            // Поддержка клавиатуры
            question.addEventListener('keydown', function(e) {
                if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    question.click();
                }
            });
        });
    }

    function initFAQToggle() {
        const toggleBtn = document.querySelector('.btn-toggle-faq');

        if (!toggleBtn) {
            return;
        }

        const hiddenItems = document.querySelectorAll('.faq-item--hidden');
        const showText = toggleBtn.querySelector('.toggle-text-show');
        const hideText = toggleBtn.querySelector('.toggle-text-hide');

        toggleBtn.addEventListener('click', function() {
            const isExpanded = toggleBtn.classList.contains('expanded');

            if (isExpanded) {
                // Скрыть дополнительные вопросы
                hiddenItems.forEach(function(item) {
                    item.classList.remove('faq-item--shown');
                    item.classList.add('faq-item--hidden');
                });
                toggleBtn.classList.remove('expanded');
                showText.style.display = 'inline';
                hideText.style.display = 'none';
            } else {
                // Показать дополнительные вопросы
                hiddenItems.forEach(function(item) {
                    item.classList.remove('faq-item--hidden');
                    item.classList.add('faq-item--shown');
                });
                toggleBtn.classList.add('expanded');
                showText.style.display = 'none';
                hideText.style.display = 'inline';
            }
        });
    }

    // Инициализация при загрузке DOM
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', function() {
            initFAQ();
            initFAQToggle();
        });
    } else {
        // DOM уже загружен
        initFAQ();
        initFAQToggle();
    }
})();
