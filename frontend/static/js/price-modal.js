(function () {
    const SELECTORS = {
        grid: '.prices-grid',
        btn: '.price-detail-btn'
    };

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init()
;
    }

    function init() {
        // Создаем модалку с галереей
        const modal = document.createElement('div');
        modal.className = 'price-modal';
        modal.innerHTML = `
            <div class="price-modal__dialog" role="dialog" aria-modal="true" aria-label="Подробная информация о цене">
                <button class="price-modal__close" aria-label="Закрыть">×</button>
                <div class="price-modal__media">
                    <button class="gallery-nav gallery-nav-prev" aria-label="Предыдущее изображение">‹</button>
                    <img id="priceModalImg" alt="">
                    <button class="gallery-nav gallery-nav-next" aria-label="Следующее изображение">›</button>
                    <div class="gallery-counter"></div>
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
            close: modal.querySelector('.price-modal__close'),
            prevBtn: modal.querySelector('.gallery-nav-prev'),
            nextBtn: modal.querySelector('.gallery-nav-next'),
            counter: modal.querySelector('.gallery-counter')
        };

        // Состояние галереи
        let galleryState = {
            images: [],
            currentIndex: 0
        };

        // ——— функции галереи ———
        function showImageAtIndex(index) {
            if (!galleryState.images || galleryState.images.length === 0) return;

            const totalImages = galleryState.images.length;
            galleryState.currentIndex = ((index % totalImages) + totalImages) % totalImages;

            const img = galleryState.images[galleryState.currentIndex];

            // Получаем путь к medium thumbnail (для галереи) с fallback к оригиналу
            const mediumThumb = getThumbnailPath(img, 'medium');
            const version = getImageVersion(img);

            // Устанавливаем medium изображение для галереи с cache-busting
            ui.mediaImg.src = `/static/uploads/${mediumThumb}?v=${version}`;
            ui.mediaImg.alt = img.alt || img.original_name || '';

            // Применяем crop стили ТОЛЬКО если thumbnails не существуют (fallback к оригиналу)
            if (!img.thumbnail_medium_path) {
                // Старое изображение без thumbnails - применяем CSS transform
                const scale = img.crop_scale || 1;
                const cropX = img.crop_x || 50;
                const cropY = img.crop_y || 50;
                const tx = (cropX - 50) * 2;
                const ty = (cropY - 50) * 2;
                ui.mediaImg.style.transform = `scale(${scale}) translate(${tx}%, ${ty}%)`;
                ui.mediaImg.style.objectFit = 'cover';
                ui.mediaImg.style.transformOrigin = 'center center';
            } else {
                // Новое изображение с thumbnails - кроп уже применен на сервере
                ui.mediaImg.style.transform = '';
                ui.mediaImg.style.objectFit = 'contain';
            }

            // Обновляем счетчик
            if (totalImages > 1) {
                ui.counter.textContent = `${galleryState.currentIndex + 1} / ${totalImages}`;
                ui.counter.style.display = 'block';
                ui.prevBtn.style.display = 'flex';
                ui.nextBtn.style.display = 'flex';
            } else {
                ui.counter.style.display = 'none';
                ui.prevBtn.style.display = 'none';
                ui.nextBtn.style.display = 'none';
            }
        }

        // Получает путь к thumbnail нужного размера с fallback
        function getThumbnailPath(img, size) {
            let thumbPath = '';
            switch (size) {
                case 'small':
                    thumbPath = img.thumbnail_small_path;
                    break;
                case 'medium':
                    thumbPath = img.thumbnail_medium_path;
                    break;
            }

            // Fallback к filename если thumbnail не существует
            if (!thumbPath) {
                return img.filename;
            }

            // Извлекаем только имя файла из полного пути
            const parts = thumbPath.split(/[/\\]/);
            return parts[parts.length - 1];
        }

        // Получает версию изображения для cache-busting
        function getImageVersion(img) {
            // Используем updated_at если доступен, иначе created_at
            if (img.updated_at) {
                // Конвертируем ISO timestamp в unix timestamp (секунды)
                return Math.floor(new Date(img.updated_at).getTime() / 1000);
            }
            if (img.created_at) {
                return Math.floor(new Date(img.created_at).getTime() / 1000);
            }
            return 0;
        }

        function showNextImage() {
            showImageAtIndex(galleryState.currentIndex + 1);
        }

        function showPrevImage() {
            showImageAtIndex(galleryState.currentIndex - 1);
        }

        // ——— делегирование клика по "Подробнее" ———
        const grid = document.querySelector(SELECTORS.grid);
        if (!grid) return;

        grid.addEventListener('click', (e) => {
            const btn = e.target.closest(SELECTORS.btn);
            if (!btn) return;

            e.preventDefault();

            // Получаем данные из data-атрибутов кнопки
            const title = btn.getAttribute('data-title') || '';
            const description = btn.getAttribute('data-description') || '';
            const priceFrom = btn.getAttribute('data-price-from') || '';
            const hasSpecs = btn.getAttribute('data-has-specs') === 'true';

            // Парсим данные изображений из data-images
            let images = [];
            try {
                const imagesData = btn.getAttribute('data-images');
                if (imagesData) {
                    images = JSON.parse(imagesData);
                }
            } catch (err) {
                console.error('Ошибка парсинга изображений:', err);
            }

            // Инициализируем галерею
            galleryState.images = images;

            // Находим индекс главного изображения (is_primary) или используем первое
            let startIndex = 0;
            if (images.length > 0) {
                const primaryIndex = images.findIndex(img => img.is_primary);
                if (primaryIndex !== -1) {
                    startIndex = primaryIndex;
                }
            }

            galleryState.currentIndex = startIndex;

            // Заполняем текстовую информацию
            ui.title.textContent = title;
            ui.price.textContent = `от ${priceFrom} ₽`;
            ui.desc.textContent = description;

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

            // Показываем первое (или главное) изображение
            showImageAtIndex(startIndex);

            openModal();
        });

        // ——— открытие/закрытие ———
        function openModal() {
            modal.classList.add('is-open');
            document.body.classList.add('modal-open');
            document.addEventListener('keydown', onKeydown);
            modal.addEventListener('click', onBackdrop);
            ui.close.addEventListener('click', closeModal);
            ui.prevBtn.addEventListener('click', showPrevImage);
            ui.nextBtn.addEventListener('click', showNextImage);
        }

        function closeModal() {
            modal.classList.remove('is-open');
            document.body.classList.remove('modal-open');
            document.removeEventListener('keydown', onKeydown);
            modal.removeEventListener('click', onBackdrop);
            ui.close.removeEventListener('click', closeModal);
            ui.prevBtn.removeEventListener('click', showPrevImage);
            ui.nextBtn.removeEventListener('click', showNextImage);
            // Чистим src, чтобы при следующем открытии не мигало на слабых сетях
            ui.mediaImg.removeAttribute('src');
        }

        function onKeydown(e) {
            if (e.key === 'Escape') {
                closeModal();
            } else if (e.key === 'ArrowLeft') {
                e.preventDefault();
                showPrevImage();
            } else if (e.key === 'ArrowRight') {
                e.preventDefault();
                showNextImage();
            }
        }

        function onBackdrop(e) {
            if (e.target === modal) closeModal();
        }
    }
})();
