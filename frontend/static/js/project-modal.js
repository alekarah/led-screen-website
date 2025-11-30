(function () {
    const SELECTORS = {
        grid: '.projects-grid',
        cardImage: '.public-project-image img',
        cardTitle: '.project-title, h3',
        cardSize: '.project-size',
        cardDesc: '.project-description',
        cardLoc:  '.project-location',
        cardTags: '.project-categories',
        btn: '.project-detail-btn'
    };

    const VIEW_ENDPOINT_BY_ID = (id) => `/api/track/project-view/${id}`;

    const VIEW_TTL_MIN = 10;

    function nowSec(){ return Math.floor(Date.now()/1000); }

    function viewedRecently(key, ttlMin){
    try{
        const raw = sessionStorage.getItem(key);
        if(!raw) return false;
        const ts = parseInt(raw, 10);
        if(Number.isNaN(ts)) return false;
        return (nowSec() - ts) < (ttlMin*60);
    }catch(_){ return false; }
    }

    function markViewedTTL(key){
    try{ sessionStorage.setItem(key, String(nowSec())); }catch(_){}
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
    return;

    function init() {
        // === модалка с галереей ===
        const modal = document.createElement('div');
        modal.className = 'project-modal';
        modal.innerHTML = `
            <div class="project-modal__dialog" role="dialog" aria-modal="true" aria-label="Информация о проекте">
                <button class="project-modal__close" aria-label="Закрыть">×</button>
                <div class="project-modal__media">
                    <button class="gallery-nav gallery-nav-prev" aria-label="Предыдущее изображение">‹</button>
                    <img id="projectModalImg" alt="">
                    <button class="gallery-nav gallery-nav-next" aria-label="Следующее изображение">›</button>
                    <div class="gallery-counter"></div>
                </div>
                <div class="project-modal__body">
                    <h3 class="project-modal__title"></h3>
                    <div class="project-modal__specs"></div>
                    <div class="project-modal__location"></div>
                    <div class="project-modal__desc"></div>
                    <a class="project-detail-btn" target="_blank" rel="noopener">Открыть изображение</a>
                </div>
            </div>
        `;
        document.body.appendChild(modal);

        const ui = {
            root: modal,
            dialog: modal.querySelector('.project-modal__dialog'),
            mediaImg: modal.querySelector('#projectModalImg'),
            title: modal.querySelector('.project-modal__title'),
            specs: modal.querySelector('.project-modal__specs'),
            desc:  modal.querySelector('.project-modal__desc'),
            loc:   modal.querySelector('.project-modal__location'),
            link:  modal.querySelector('.project-modal__body .project-detail-btn'),
            close: modal.querySelector('.project-modal__close'),
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

            // Обновляем ссылку на оригинал для кнопки "Открыть" (максимальное качество)
            ui.link.href = `/static/uploads/${img.filename}?v=${version}`;

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

    // Автооткрытие модалки по hash в URL (при переходе с главной)
    if (window.location.hash) {
        const slug = window.location.hash.substring(1); // убираем #
        const targetBtn = document.querySelector(`[data-project-slug="${slug}"]`);
        if (targetBtn) {
            // Небольшая задержка для полной загрузки страницы
            setTimeout(() => {
                targetBtn.click();
                // Убираем hash из URL после открытия модалки
                history.replaceState(null, null, ' ');
            }, 300);
        }
    }

    grid.addEventListener('click', (e) => {
        const btn = e.target.closest(SELECTORS.btn);
        if (!btn) return;

        e.preventDefault();

        const card = btn.closest('.public-project-card');
        if (!card) return;

        const titleEl = card.querySelector(SELECTORS.cardTitle);
        const sizeEl  = card.querySelector(SELECTORS.cardSize);

        // Получаем данные из data-атрибутов кнопки
        const pixelPitch = btn.getAttribute('data-pixel-pitch') || '';
        const location = btn.getAttribute('data-location') || '';
        const description = btn.getAttribute('data-description') || '';

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

        // Если нет изображений в data, используем изображение из карточки
        if (!images || images.length === 0) {
            const imgEl = card.querySelector(SELECTORS.cardImage);
            const imgSrc = imgEl?.getAttribute('src') || '';
            const imgAlt = imgEl?.getAttribute('alt') || '';

            if (imgSrc) {
                const filename = imgSrc.split('/').pop();
                images = [{
                    filename: filename,
                    alt: imgAlt,
                    crop_x: 50,
                    crop_y: 50,
                    crop_scale: 1
                }];
            }
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

        const projectId = btn.getAttribute('data-project-id');
        const key = `pview:${projectId}`;
        if (!viewedRecently(key, VIEW_TTL_MIN)) {
            try { fetch(VIEW_ENDPOINT_BY_ID(projectId), { method: 'POST' }); } catch(_) {}
            markViewedTTL(key);
        }

        // Заполняем текстовую информацию
        ui.title.textContent = titleEl?.textContent?.trim() ?? '';

        // Объединяем размер и пиксели в одну строку через разделитель
        const size = sizeEl?.textContent?.trim() ?? '';
        const specs = [size, pixelPitch].filter(Boolean).join('  •  ');
        ui.specs.textContent = specs;

        ui.loc.textContent   = location;
        ui.desc.textContent  = description;

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
            if (e.target === modal) closeModal(); // клик по подложке
        }
    }
})();