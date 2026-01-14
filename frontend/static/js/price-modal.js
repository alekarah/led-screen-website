(function () {
    const SELECTORS = {
        grid: '.prices-grid',
        btn: '.price-detail-btn'
    };

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    function init() {
        // Создаем модалку
        const modal = document.createElement('div');
        modal.className = 'price-modal';
        modal.innerHTML = `
            <div class="price-modal__dialog" role="dialog" aria-modal="true" aria-label="Подробная информация о цене">
                <button class="price-modal__close" aria-label="Закрыть">×</button>
                <div class="price-modal__media">
                    <img id="priceModalImg" alt="">
                </div>
                <div class="price-modal__body">
                    <div class="price-modal__content">
                        <h3 class="price-modal__title"></h3>
                        <div class="price-modal__price"></div>
                        <div class="price-modal__desc"></div>
                        <div class="price-modal__specs"></div>
                    </div>
                    <div class="price-modal__footer">
                        <a href="/contact" class="btn">Отправить заявку</a>
                    </div>
                </div>
            </div>
        `;
        document.body.appendChild(modal);

        const ui = {
            root: modal,
            dialog: modal.querySelector('.price-modal__dialog'),
            mediaImg: modal.querySelector('#priceModalImg'),
            title: modal.querySelector('.price-modal__title'),
            price: modal.querySelector('.price-modal__price'),
            desc: modal.querySelector('.price-modal__desc'),
            specs: modal.querySelector('.price-modal__specs'),
            close: modal.querySelector('.price-modal__close')
        };

        // Делегирование клика по "Подробнее"
        const grid = document.querySelector(SELECTORS.grid);
        if (!grid) return;

        grid.addEventListener('click', (e) => {
            const btn = e.target.closest(SELECTORS.btn);
            if (!btn) return;

            e.preventDefault();

            // Получаем данные из data-атрибутов
            const title = btn.getAttribute('data-title') || '';
            const description = btn.getAttribute('data-description') || '';
            const priceFrom = btn.getAttribute('data-price-from') || '';
            const imagePath = btn.getAttribute('data-image-path') || '';
            const hasSpecs = btn.getAttribute('data-has-specs') === 'true';

            // Парсим характеристики
            let specs = [];
            try {
                const specsData = btn.getAttribute('data-specs');
                if (specsData) {
                    specs = JSON.parse(specsData);
                }
            } catch (err) {
                console.error('Ошибка парсинга характеристик:', err);
            }

            // Заполняем модалку
            ui.title.textContent = title;
            ui.price.textContent = `от ${priceFrom} ₽`;
            ui.desc.textContent = description;

            // Устанавливаем изображение
            if (imagePath) {
                ui.mediaImg.src = imagePath;
                ui.mediaImg.alt = title;
            } else {
                ui.mediaImg.src = '/static/images/placeholder.jpg';
                ui.mediaImg.alt = 'Нет изображения';
            }

            // Заполняем характеристики
            if (hasSpecs && specs.length > 0) {
                let specsHTML = '<h4>Характеристики</h4>';

                specs.forEach((group, index) => {
                    // Group передается с заглавной буквы из Go handler (GroupedSpec)
                    const groupName = group.Group || '';
                    const groupSpecs = group.Specs || [];

                    if (!groupName || groupSpecs.length === 0) return;

                    const groupId = `spec-group-${index}`;

                    specsHTML += `
                        <div class="spec-accordion-item">
                            <button class="spec-accordion-header"
                                    aria-expanded="false"
                                    aria-controls="${groupId}">
                                <span>${groupName}</span>
                                <svg class="spec-accordion-icon" width="16" height="16" viewBox="0 0 16 16" fill="none">
                                    <path d="M4 6L8 10L12 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                                </svg>
                            </button>
                            <div class="spec-accordion-content" id="${groupId}" hidden>
                                <table class="price-modal__specs-table">
                                    <tbody>
                    `;

                    groupSpecs.forEach(spec => {
                        // JSON теги в Go модели используют snake_case
                        const key = spec.spec_key || '';
                        const value = spec.spec_value || '';
                        specsHTML += `
                            <tr>
                                <td class="spec-key">${key}</td>
                                <td class="spec-value">${value}</td>
                            </tr>
                        `;
                    });

                    specsHTML += `
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    `;
                });

                ui.specs.innerHTML = specsHTML;

                // Добавляем обработчики для аккордеона
                const accordionHeaders = ui.specs.querySelectorAll('.spec-accordion-header');
                accordionHeaders.forEach(header => {
                    header.addEventListener('click', () => {
                        const content = header.nextElementSibling;
                        const isExpanded = header.getAttribute('aria-expanded') === 'true';

                        header.setAttribute('aria-expanded', !isExpanded);
                        content.hidden = isExpanded;
                    });
                });
            } else {
                ui.specs.innerHTML = '';
            }

            openModal();
        });

        // Открытие/закрытие модалки
        function openModal() {
            modal.classList.add('is-open');
            document.body.classList.add('modal-open');
            document.addEventListener('keydown', onKeydown);
            modal.addEventListener('click', onBackdrop);
            ui.close.addEventListener('click', closeModal);
        }

        function closeModal() {
            modal.classList.remove('is-open');
            document.body.classList.remove('modal-open');
            document.removeEventListener('keydown', onKeydown);
            modal.removeEventListener('click', onBackdrop);
            ui.close.removeEventListener('click', closeModal);
            ui.mediaImg.removeAttribute('src');
        }

        function onKeydown(e) {
            if (e.key === 'Escape') {
                closeModal();
            }
        }

        function onBackdrop(e) {
            if (e.target === modal) closeModal();
        }
    }
})();
