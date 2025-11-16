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
                var hasOtherActive = false;

                // Проверяем, есть ли другие открытые элементы
                faqItems.forEach(function(otherItem) {
                    if (otherItem !== item && otherItem.classList.contains('active')) {
                        hasOtherActive = true;
                        otherItem.classList.remove('active');
                    }
                });

                // Если был другой открытый элемент и текущий закрыт,
                // добавляем небольшую задержку для плавности
                if (hasOtherActive && !isCurrentlyActive) {
                    setTimeout(function() {
                        item.classList.add('active');
                    }, 150);
                } else {
                    // Просто переключаем текущий элемент
                    item.classList.toggle('active');
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

    // Инициализация при загрузке DOM
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initFAQ);
    } else {
        // DOM уже загружен
        initFAQ();
    }
})();
