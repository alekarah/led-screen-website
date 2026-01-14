// Аккордеон для характеристик на странице цен

document.addEventListener('DOMContentLoaded', () => {
    const toggleButtons = document.querySelectorAll('.specs-toggle');

    toggleButtons.forEach(button => {
        button.addEventListener('click', () => {
            const isExpanded = button.getAttribute('aria-expanded') === 'true';
            const contentId = button.getAttribute('aria-controls');
            const content = document.getElementById(contentId);

            if (!content) return;

            // Переключаем состояние
            button.setAttribute('aria-expanded', !isExpanded);
            content.hidden = isExpanded;
        });
    });
});
