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
        // === модалка ===
        const modal = document.createElement('div');
        modal.className = 'project-modal';
        modal.innerHTML = `
            <div class="project-modal__dialog" role="dialog" aria-modal="true" aria-label="Информация о проекте">
                <button class="project-modal__close" aria-label="Закрыть">×</button>
                <div class="project-modal__media">
                    <img id="projectModalImg" alt="">
                </div>
                <div class="project-modal__body">
                    <h3 class="project-modal__title"></h3>
                    <div class="project-modal__size"></div>
                    <div class="project-modal__desc"></div>
                    <div class="project-modal__location"></div>
                    <div class="project-modal__tags"></div>
                    <a class="project-detail-btn" target="_blank" rel="noopener" style="margin-top:.35rem;display:inline-block;">Открыть изображение</a>
                </div>
            </div>
        `;
        document.body.appendChild(modal);

        const ui = {
            root: modal,
            dialog: modal.querySelector('.project-modal__dialog'),
            mediaImg: modal.querySelector('#projectModalImg'),
            title: modal.querySelector('.project-modal__title'),
            size:  modal.querySelector('.project-modal__size'),
            desc:  modal.querySelector('.project-modal__desc'),
            loc:   modal.querySelector('.project-modal__location'),
            tags:  modal.querySelector('.project-modal__tags'),
            link:  modal.querySelector('.project-modal__body .project-detail-btn'),
            close: modal.querySelector('.project-modal__close')
        };

        // ——— утилита: скопировать кроп-стили из карточки в модалку ———
        function copyCropStyles(fromImg, toImg) {
            // 1) пробуем взять inline transform/object-position/object-fit (как на карточках/главной)
            const style = fromImg.getAttribute('style') || '';

            const mTransform = style.match(/transform\s*:\s*([^;]+)/i);
            const mObjPos   = style.match(/object-position\s*:\s*([^;]+)/i);
            const mObjFit   = style.match(/object-fit\s*:\s*([^;]+)/i);

            if (mTransform) toImg.style.transform = mTransform[1].trim();
            else            toImg.style.removeProperty('transform');

            toImg.style.objectPosition = (mObjPos ? mObjPos[1].trim() : 'center center');
            toImg.style.objectFit      = (mObjFit ? mObjFit[1].trim() : 'cover');

            // 2) запасной путь: если рендеришь data-* (cropScale/cropX/cropY)
            if (!mTransform) {
            const sx = parseFloat(fromImg.dataset.cropScale || '');
            const cx = parseFloat(fromImg.dataset.cropX || '');
            const cy = parseFloat(fromImg.dataset.cropY || '');
            if (Number.isFinite(sx) && Number.isFinite(cx) && Number.isFinite(cy)) {
                const tx = (cx - 50) * 2;
                const ty = (cy - 50) * 2;
                toImg.style.transform = `scale(${sx}) translate(${tx}%, ${ty}%)`;
                toImg.style.objectPosition = 'center center';
                toImg.style.objectFit = 'cover';
            }
        }
    }

    // ——— делегирование клика по "Подробнее" ———
    const grid = document.querySelector(SELECTORS.grid);
    if (!grid) return;

    grid.addEventListener('click', (e) => {
        const btn = e.target.closest(SELECTORS.btn);
        if (!btn) return;

        e.preventDefault();

        const card = btn.closest('.public-project-card');
        if (!card) return;

        const imgEl  = card.querySelector(SELECTORS.cardImage);
        const titleEl = card.querySelector(SELECTORS.cardTitle);
        const sizeEl  = card.querySelector(SELECTORS.cardSize);
        const descEl  = card.querySelector(SELECTORS.cardDesc);
        const locEl   = card.querySelector(SELECTORS.cardLoc);
        const tagsEl  = card.querySelector(SELECTORS.cardTags);

        const imgSrc = imgEl?.getAttribute('src') || '';
        const imgAlt = imgEl?.getAttribute('alt') || (titleEl?.textContent?.trim() ?? 'Изображение проекта');

        const projectId = btn.getAttribute('data-project-id');
        const key = `pview:${projectId}`;
        if (!viewedRecently(key, VIEW_TTL_MIN)) {
        try { fetch(VIEW_ENDPOINT_BY_ID(projectId), { method: 'POST' }); } catch(_) {}
        markViewedTTL(key);
        }

        // картинка
        ui.mediaImg.src = imgSrc;
        ui.mediaImg.alt = imgAlt;
        copyCropStyles(imgEl, ui.mediaImg);

        // текст
        ui.title.textContent = titleEl?.textContent?.trim() ?? '';
        ui.size.textContent  = sizeEl?.textContent?.trim() ?? '';
        ui.desc.textContent  = descEl?.textContent?.trim() ?? '';
        ui.loc.textContent   = locEl?.textContent?.trim() ?? '';
        ui.tags.innerHTML    = tagsEl?.innerHTML ?? '';

        ui.link.href = imgSrc;

        openModal();
    });

        // ——— открытие/закрытие ———
        function openModal() {
            modal.classList.add('is-open');
            document.body.classList.add('modal-open');
            document.addEventListener('keydown', onEsc);
            modal.addEventListener('click', onBackdrop);
            ui.close.addEventListener('click', closeModal);
        }

        function closeModal() {
            modal.classList.remove('is-open');
            document.body.classList.remove('modal-open');
            document.removeEventListener('keydown', onEsc);
            modal.removeEventListener('click', onBackdrop);
            ui.close.removeEventListener('click', closeModal);
            // Чистим src, чтобы при следующем открытии не мигало на слабых сетях
            ui.mediaImg.removeAttribute('src');
        }

        function onEsc(e) {
            if (e.key === 'Escape') closeModal();
        }
        function onBackdrop(e) {
            if (e.target === modal) closeModal(); // клик по подложке
        }
    }
})();