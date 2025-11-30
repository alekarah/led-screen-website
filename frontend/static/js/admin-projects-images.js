// Работа с изображениями проектов

// --- публичный API ---
window.fillProjectImages = fillProjectImages;
window.initImagePreview  = initImagePreview;

// оставляем на всякий случай:
window.handleImageError  = handleImageError;

// =============================
// РЕНДЕР КОЛЛЕКЦИИ ИЗОБРАЖЕНИЙ
// =============================
function fillProjectImages(images) {
    const root = document.getElementById('project_images');
    if (!root) {
        console.warn('Контейнер изображений не найден');
        return;
    }

    // Сообщения о пустом состоянии
    if (!Array.isArray(images)) {
        root.innerHTML = `<p class="no-images">Изображения не загружены</p>`;
        return;
    }
    if (images.length === 0) {
        root.innerHTML = `<p class="no-images">У проекта пока нет изображений</p>`;
        return;
    }

    // Рендер через DocumentFragment (без лишних reflow)
    root.innerHTML = '';
    const frag = document.createDocumentFragment();
    images.forEach(img => frag.appendChild(buildImageItem(img)));
    root.appendChild(frag);

    // Включаем делегирование событий один раз
    enableImagesDelegation(root);
}

// Строим карточку изображения как DOM‑узлы
function buildImageItem(image) {
    const item = document.createElement('div');
    item.className = 'image-item';
    item.dataset.imageId = image.id;

    const previewWrap = document.createElement('div');
    previewWrap.className = 'image-preview-container';

    const img = document.createElement('img');
    // Используем small thumbnail для превью, fallback к оригиналу
    const previewSrc = image.thumbnail_small_path || image.filename;
    // Извлекаем только имя файла из пути (на случай если в БД хранится полный путь)
    const filename = previewSrc.split(/[/\\]/).pop();
    // Добавляем агрессивный cache-busting для перезагрузки после кроппинга
    // Используем ID изображения + timestamp + random для гарантии уникальности
    const cacheBuster = `v=${image.id}_${Date.now()}_${Math.random().toString(36).substr(2, 5)}`;
    img.src = `/static/uploads/${encodeURIComponent(filename)}?${cacheBuster}`;
    img.alt = escapeHtml(image.alt || image.original_name || '');
    img.dataset.originalFilename = image.filename; // для fallback
    img.onerror = () => handleImageError(img);

    const loading = document.createElement('div');
    loading.className = 'image-loading hidden';
    loading.innerHTML = `<div class="loading-spinner"></div>`;

    previewWrap.appendChild(img);
    previewWrap.appendChild(loading);

    // info
    const info = document.createElement('div');
    info.className = 'image-info';
    const nameSpan = document.createElement('span');
    nameSpan.className = 'image-name';
    nameSpan.title = escapeHtml(image.original_name || image.filename);
    nameSpan.textContent = truncateText(image.original_name || image.filename, 15);
    const sizeSpan = document.createElement('span');
    sizeSpan.className = 'image-size';
    sizeSpan.textContent = formatFileSize(image.file_size || 0);
    info.appendChild(nameSpan);
    info.appendChild(sizeSpan);

    // controls
    const controls = document.createElement('div');
    controls.className = 'image-controls';

    const cropBtn = document.createElement('button');
    cropBtn.className = 'crop-image';
    cropBtn.type = 'button';
    cropBtn.title = 'Настроить отображение';
    cropBtn.setAttribute('aria-label', 'Настроить кроппинг изображения');
    cropBtn.dataset.action = 'crop';
    // данные для редактора
    cropBtn.dataset.imageId = image.id;
    cropBtn.dataset.filename = image.filename;
    cropBtn.dataset.cropX = image.crop_x || 50;
    cropBtn.dataset.cropY = image.crop_y || 50;
    cropBtn.dataset.cropScale = image.crop_scale || 1;
    cropBtn.textContent = '✂️';

    // Кнопка "Сделать главным" или индикатор главного изображения
    const primaryBtn = document.createElement('button');
    primaryBtn.type = 'button';
    primaryBtn.dataset.imageId = image.id;

    if (image.is_primary) {
        // Если это главное изображение - показываем индикатор (желтая с черной звездой)
        primaryBtn.className = 'primary-image-indicator';
        primaryBtn.title = 'Главное изображение проекта';
        primaryBtn.setAttribute('aria-label', 'Главное изображение проекта');
        primaryBtn.textContent = '★';
        primaryBtn.disabled = true;
    } else {
        // Если не главное - показываем кнопку для установки (серая с черной звездой)
        primaryBtn.className = 'set-primary-image';
        primaryBtn.title = 'Сделать главным изображением';
        primaryBtn.setAttribute('aria-label', 'Сделать главным изображением');
        primaryBtn.dataset.action = 'set-primary';
        primaryBtn.textContent = '★';
    }

    const delBtn = document.createElement('button');
    delBtn.className = 'delete-image';
    delBtn.type = 'button';
    delBtn.title = 'Удалить изображение';
    delBtn.setAttribute('aria-label', 'Удалить изображение');
    delBtn.dataset.action = 'delete';
    delBtn.dataset.imageId = image.id;
    delBtn.textContent = '×';

    controls.appendChild(cropBtn);
    controls.appendChild(primaryBtn);
    controls.appendChild(delBtn);

    item.appendChild(previewWrap);
    item.appendChild(info);
    item.appendChild(controls);

    return item;
}

// Делегирование событий на контейнере, вместо inline onclick
let __imagesDelegationBound = false;
function enableImagesDelegation(root) {
    if (__imagesDelegationBound) return;
    __imagesDelegationBound = true;

    root.addEventListener('click', async (e) => {
        const btn = e.target.closest('button[data-action]');
        if (!btn) return;

        const imageId = Number(btn.dataset.imageId);
        const action = btn.dataset.action;

        if (action === 'delete') {
            await deleteImage(imageId);
            return;
        }

        if (action === 'crop') {
            const x = Number(btn.dataset.cropX);
            const y = Number(btn.dataset.cropY);
            const scale = Number(btn.dataset.cropScale);
            const filename = btn.dataset.filename;
            if (typeof openCropEditor === 'function') {
                openCropEditor(imageId, filename, x, y, scale);
            }
            return;
        }

        if (action === 'set-primary') {
            await setPrimaryImage(imageId);
            return;
        }
    });
}

// =============================
// ДЕЙСТВИЯ С КАРТИНКАМИ
// =============================
async function deleteImage(imageId) {
    if (!imageId) {
        showAdminMessage('Некорректный ID изображения', 'error');
        return;
    }

    confirmAction('Удалить это изображение?', async () => {
        showImageLoading(imageId, true);
        try {
        await deleteData(`/admin/images/${imageId}`);
        showAdminMessage('Изображение удалено', 'success');
        removeImageFromDOM(imageId);

        // синхронизация списка
        const projectId = document.getElementById('edit_project_id')?.value;
        if (projectId) {
            setTimeout(async () => {
            try {
                const data = await fetchData(`/admin/projects/${projectId}?_=${Date.now()}`);
                fillProjectImages(data.project.images || []);
            } catch (err) {
                console.error('Ошибка обновления списка изображений:', err);
            }
            }, 300);
        }
        } catch (err) {
        console.error('Ошибка удаления изображения:', err);
        showAdminMessage('Ошибка при удалении: ' + err.message, 'error');
        } finally {
        showImageLoading(imageId, false);
        }
    });
}

function removeImageFromDOM(imageId) {
    const el = document.querySelector(`[data-image-id="${imageId}"]`);
    if (!el) return;
    el.style.animation = 'fadeOut .2s ease-out';
    setTimeout(() => el.remove(), 200);
}

async function setPrimaryImage(imageId) {
    if (!imageId) {
        showAdminMessage('Некорректный ID изображения', 'error');
        return;
    }

    try {
        showImageLoading(imageId, true);

        const response = await fetch(`/admin/images/${imageId}/set-primary`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Ошибка установки главного изображения');
        }

        const result = await response.json();
        showAdminMessage(result.message || 'Главное изображение установлено', 'success');

        // Перезагружаем список изображений для обновления индикаторов
        const projectId = document.getElementById('edit_project_id')?.value;
        if (projectId) {
            const projectData = await fetchData(`/admin/projects/${projectId}?_=${Date.now()}`);
            fillProjectImages(projectData.project.images || []);
        }
    } catch (err) {
        console.error('Ошибка установки главного изображения:', err);
        showAdminMessage('Ошибка: ' + err.message, 'error');
    } finally {
        showImageLoading(imageId, false);
    }
}

// =============================
// ПРЕВЬЮ ПЕРЕД ЗАГРУЗКОЙ (форма в «Редактировать»)
// =============================
function initImagePreview() {
    // только для инпутов именно в форме редактирования
    const input = document.querySelector('#editUploadForm input[type="file"][accept*="image"]');
    if (!input || input.dataset.bound === '1') return;
    input.dataset.bound = '1';

    input.addEventListener('change', () => {
        const container = getOrCreatePreviewContainer(input);
        container.innerHTML = '';
        const files = Array.from(input.files || []);
        files.forEach(file => {
        if (!file.type.startsWith('image/')) return;
        const reader = new FileReader();
        reader.onload = (e) => {
            const wrap = document.createElement('div');
            wrap.className = 'upload-preview';
            wrap.innerHTML = `
            <img src="${e.target.result}" alt="Превью">
            <div class="preview-info">
                <span class="preview-name">${truncateText(file.name, 12)}</span>
                <span class="preview-size">${formatFileSize(file.size)}</span>
            </div>
            <button type="button" class="remove-preview" aria-label="Удалить превью">×</button>
            `;
            container.appendChild(wrap);

            // удалить конкретный файл из списка input.files
            wrap.querySelector('.remove-preview').addEventListener('click', () => {
            const dt = new DataTransfer();
            Array.from(input.files).forEach(f => { if (f !== file) dt.items.add(f); });
            input.files = dt.files;
            wrap.remove();
            });
        };
        reader.readAsDataURL(file);
        });
    });
}

// Получение или создание контейнера для превью
function getOrCreatePreviewContainer(input) {
    let container = input.parentNode.querySelector('.image-preview-container');
    
    if (!container) {
        container = document.createElement('div');
        container.className = 'image-preview-container';
        input.parentNode.appendChild(container);
    }
    
    return container;
}

// =============================
// УТИЛИТЫ
// =============================
function handleImageError(img) {
    // Предотвращаем бесконечный цикл - если уже пробовали оригинал, показываем заглушку
    if (img.dataset.triedOriginal === 'true') {
        img.onerror = null; // убираем обработчик чтобы не было цикла
        img.style.display = 'none'; // скрываем сломанное изображение
        img.alt = 'Изображение не найдено';
        return;
    }

    // Пробуем загрузить оригинал как fallback
    const originalFilename = img.dataset.originalFilename;
    if (originalFilename) {
        img.dataset.triedOriginal = 'true';
        img.src = `/static/uploads/${encodeURIComponent(originalFilename)}`;
    } else {
        img.onerror = null;
        img.style.display = 'none';
    }
}

function showImageLoading(imageId, show) {
    const el = document.querySelector(`[data-image-id="${imageId}"]`);
    if (!el) return;
    const spinner = el.querySelector('.image-loading');
    const controls = el.querySelector('.image-controls');
    if (spinner) spinner.classList.toggle('hidden', !show);
    if (controls) controls.style.opacity = show ? '0.5' : '1';
}

function truncateText(text, max) {
    if (!text || text.length <= max) return text || '';
    return text.slice(0, max) + '…';
}

function formatFileSize(bytes) {
    if (!bytes) return '0 B';
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round((bytes / Math.pow(1024, i)) * 100) / 100 + ' ' + sizes[i];
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text || '';
    return div.innerHTML;
}
