// Работа с множественными изображениями позиций прайса (аналогично admin-projects-images.js)

// --- публичный API ---
window.displayPriceImages = displayPriceImages;
window.initPriceImageUploadForm = initPriceImageUploadForm;
window.initPriceImagePreview = initPriceImagePreview;

// =============================
// РЕНДЕР КОЛЛЕКЦИИ ИЗОБРАЖЕНИЙ
// =============================
function displayPriceImages(images) {
    const root = document.getElementById('price_images');
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
        root.innerHTML = `<p class="no-images">У позиции пока нет изображений</p>`;
        return;
    }

    // Рендер через DocumentFragment (без лишних reflow)
    root.innerHTML = '';
    const frag = document.createDocumentFragment();
    images.forEach(img => frag.appendChild(buildPriceImageItem(img)));
    root.appendChild(frag);

    // Включаем делегирование событий один раз
    enableImagesDelegation(root);
}

// Строим карточку изображения как DOM‑узлы
function buildPriceImageItem(image) {
    const item = document.createElement('div');
    item.className = 'image-item';
    item.dataset.imageId = image.id;

    const previewWrap = document.createElement('div');
    previewWrap.className = 'image-preview-container';

    const img = document.createElement('img');
    // Используем small thumbnail для превью, fallback к file_path, потом к filename
    const previewSrc = image.thumbnail_small_path || image.file_path || image.filename;
    const filename = previewSrc.split(/[/\\]/).pop();
    // Агрессивный cache-busting для перезагрузки после кроппинга
    const cacheBuster = `v=${image.id}_${Date.now()}_${Math.random().toString(36).substr(2, 5)}`;
    img.src = `/static/uploads/${encodeURIComponent(filename)}?${cacheBuster}`;
    img.alt = escapeHtml(image.alt || image.original_name || '');
    img.dataset.originalFilename = image.filename;
    img.onerror = () => handlePriceImageError(img);

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
        primaryBtn.title = 'Главное изображение позиции';
        primaryBtn.setAttribute('aria-label', 'Главное изображение позиции');
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

// Делегирование событий на контейнере
let __imagesDelegationBound = false;
function enableImagesDelegation(root) {
    if (__imagesDelegationBound) return;
    __imagesDelegationBound = true;

    root.addEventListener('click', async (e) => {
        const btn = e.target.closest('button[data-action]');
        if (!btn) return;

        const action = btn.dataset.action;
        const imageId = btn.dataset.imageId;

        if (action === 'crop') {
            openPriceCropEditor(
                imageId,
                btn.dataset.filename,
                parseFloat(btn.dataset.cropX) || 50,
                parseFloat(btn.dataset.cropY) || 50,
                parseFloat(btn.dataset.cropScale) || 1.0
            );
        } else if (action === 'set-primary') {
            await setPrimaryPriceImage(imageId);
        } else if (action === 'delete') {
            await deletePriceImage(imageId);
        }
    });
}

// Обработка ошибки загрузки изображения
function handlePriceImageError(img) {
    if (img.dataset.triedOriginal === 'true') {
        img.onerror = null;
        img.style.display = 'none';
        img.alt = 'Изображение не найдено';
        return;
    }

    const originalFilename = img.dataset.originalFilename;
    if (originalFilename) {
        img.dataset.triedOriginal = 'true';
        const filename = originalFilename.split(/[/\\]/).pop();
        img.src = `/static/uploads/${encodeURIComponent(filename)}`;
    } else {
        img.onerror = null;
        img.style.display = 'none';
    }
}

// ============== УСТАНОВКА ГЛАВНОГО ИЗОБРАЖЕНИЯ ==============

async function setPrimaryPriceImage(imageId) {
    try {
        const response = await fetch(`/admin/prices/images/${imageId}/set-primary`, {
            method: 'POST'
        });

        const data = await response.json();

        if (data.success) {
            showAdminMessage('Главное изображение установлено', 'success');

            // Обновляем отображение изображений
            const priceId = document.getElementById('editPriceId')?.value;
            if (priceId) {
                await updatePriceImages(priceId);
            }
        } else {
            showAdminMessage(data.error || 'Ошибка установки главного изображения', 'error');
        }
    } catch (error) {
        console.error('Ошибка установки главного изображения:', error);
        showAdminMessage('Ошибка установки главного изображения', 'error');
    }
}

// ============== РЕДАКТОР КРОППИНГА ==============

let currentPriceCropEditor = null;

function openPriceCropEditor(imageId, imagePath, cropX, cropY, cropScale) {
    // Используем общий редактор кроппинга из crop-editor-api.js
    if (typeof openCropEditor !== 'function') {
        showAdminMessage('Редактор кроппинга не загружен', 'error');
        return;
    }

    const filename = imagePath.split(/[/\\]/).pop();

    // Сохраняем контекст
    currentPriceCropEditor = {
        imageId: imageId
    };

    // Открываем редактор кроппинга
    openCropEditor(imageId, filename, cropX, cropY, cropScale);

    // Переопределяем saveCrop() чтобы сохранять в наш endpoint
    window.originalSaveCrop = window.saveCrop;
    window.saveCrop = async function() {
        const cropXSlider = document.getElementById('cropX');
        const cropYSlider = document.getElementById('cropY');
        const cropScaleSlider = document.getElementById('cropScale');

        if (!cropXSlider || !cropYSlider || !cropScaleSlider) {
            console.error('Не найдены элементы управления кроппингом');
            return;
        }

        const cropX = parseFloat(cropXSlider.value);
        const cropY = parseFloat(cropYSlider.value);
        const cropScale = parseFloat(cropScaleSlider.value);

        try {
            const response = await fetch(`/admin/prices/images/${currentPriceCropEditor.imageId}/crop`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    crop_x: cropX,
                    crop_y: cropY,
                    crop_scale: cropScale
                })
            });

            const result = await response.json();

            if (response.ok && result.success) {
                showCropMessage('Настройки изображения сохранены!', 'success');

                // Закрываем модальное окно
                setTimeout(() => {
                    closeCropModal();

                    // Обновляем изображения в форме редактирования
                    setTimeout(() => {
                        const priceId = document.getElementById('editPriceId')?.value;
                        if (priceId) {
                            updatePriceImages(priceId);
                        }
                    }, 300);
                }, 500);

            } else {
                throw new Error(result.error || 'Ошибка сохранения');
            }

        } catch (error) {
            console.error('Ошибка сохранения кроппинга:', error);
            showCropMessage('Ошибка при сохранении: ' + error.message, 'error');
        }
    };
}

// Обновление изображений позиции после кроппинга
async function updatePriceImages(priceId) {
    try {
        const response = await fetch(`/admin/prices/${priceId}`);
        const data = await response.json();

        if (response.ok && data.price_item) {
            displayPriceImages(data.price_item.images || []);
        }
    } catch (error) {
        console.error('Ошибка обновления изображений:', error);
    }
}

// ============== УДАЛЕНИЕ ИЗОБРАЖЕНИЯ ==============

async function deletePriceImage(imageId) {
    if (!confirm('Вы уверены, что хотите удалить изображение? Это действие нельзя отменить.')) {
        return;
    }

    try {
        const response = await fetch(`/admin/prices/images/${imageId}`, {
            method: 'DELETE'
        });

        const data = await response.json();

        if (data.success) {
            showAdminMessage('Изображение удалено', 'success');

            // Обновляем отображение изображений
            const priceId = document.getElementById('editPriceId')?.value;
            if (priceId) {
                await updatePriceImages(priceId);
            }
        } else {
            showAdminMessage(data.error || 'Ошибка удаления изображения', 'error');
        }
    } catch (error) {
        console.error('Ошибка удаления изображения:', error);
        showAdminMessage('Ошибка удаления изображения', 'error');
    }
}

// ============== ЗАГРУЗКА ИЗОБРАЖЕНИЙ ==============

function initPriceImageUploadForm() {
    const form = document.getElementById('editPriceImageForm');
    if (!form || form.dataset.bound === '1') return;
    form.dataset.bound = '1';

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const priceId = document.getElementById('editPriceId')?.value;
        if (!priceId) {
            showAdminMessage('ID позиции не найден', 'error');
            return;
        }

        const fileInput = document.getElementById('editPriceImages');
        if (!fileInput || !fileInput.files || !fileInput.files.length) {
            showAdminMessage('Выберите файлы для загрузки', 'error');
            return;
        }

        // Создаем FormData с изображениями
        const formData = new FormData();
        formData.append('price_item_id', priceId);

        // Добавляем все выбранные файлы
        Array.from(fileInput.files).forEach(file => {
            formData.append('images', file);
        });

        const spinner = document.getElementById('priceUploadSpinner');
        if (spinner) spinner.classList.remove('hidden');

        try {
            const response = await fetch('/admin/prices/upload-images', {
                method: 'POST',
                body: formData
            });

            const data = await response.json();

            if (response.ok && data.images) {
                showAdminMessage(`Загружено ${data.images.length} изображений`, 'success');

                // Очищаем поле выбора файла
                fileInput.value = '';

                // Удаляем контейнер превью полностью
                const previewContainer = fileInput.parentNode.querySelector('.image-preview-container');
                if (previewContainer) {
                    previewContainer.remove();
                }

                // Обновляем отображение изображений
                await updatePriceImages(priceId);
            } else {
                showAdminMessage(data.error || 'Ошибка загрузки изображений', 'error');
            }
        } catch (error) {
            console.error('Ошибка загрузки изображений:', error);
            showAdminMessage('Ошибка загрузки изображений', 'error');
        } finally {
            if (spinner) spinner.classList.add('hidden');
        }
    });
}

// ============== ПРЕВЬЮ ПЕРЕД ЗАГРУЗКОЙ ==============

function initPriceImagePreview() {
    const input = document.querySelector('#editPriceImageForm input[type="file"][accept*="image"]');
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

                // удалить файл из списка
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

// ============== УТИЛИТЫ ==============

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
    div.textContent = text;
    return div.innerHTML;
}

// ============== ИНИЦИАЛИЗАЦИЯ ==============

document.addEventListener('DOMContentLoaded', () => {
    initPriceImageUploadForm();
    initPriceImagePreview();
});
