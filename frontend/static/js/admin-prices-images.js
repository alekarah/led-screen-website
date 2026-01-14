// Работа с изображениями позиций прайса (аналогично admin-projects-images.js)

// ============== ОТОБРАЖЕНИЕ ИЗОБРАЖЕНИЯ В МОДАЛКЕ РЕДАКТИРОВАНИЯ ==============

function displayPriceImage(priceItem) {
    const root = document.getElementById('price_images');
    if (!root) {
        console.warn('Контейнер изображения не найден');
        return;
    }

    // Если изображения нет - показываем сообщение
    if (!priceItem.image_path) {
        root.innerHTML = `<p class="no-images">У позиции пока нет изображения</p>`;
        return;
    }

    // Рендер карточки изображения
    root.innerHTML = '';
    const imageItem = buildPriceImageItem(priceItem);
    root.appendChild(imageItem);
}

// Строим карточку изображения как DOM‑узлы (аналогично buildImageItem в проектах)
function buildPriceImageItem(priceItem) {
    const item = document.createElement('div');
    item.className = 'image-item';
    item.dataset.priceId = priceItem.id;

    const previewWrap = document.createElement('div');
    previewWrap.className = 'image-preview-container';

    const img = document.createElement('img');
    // Используем small thumbnail для превью, fallback к оригиналу
    const previewSrc = priceItem.thumbnail_small_path || priceItem.image_path;
    // Извлекаем только имя файла из пути
    const filename = previewSrc.split(/[/\\]/).pop();
    // Агрессивный cache-busting для перезагрузки после кроппинга
    const cacheBuster = `v=${priceItem.id}_${Date.now()}_${Math.random().toString(36).substr(2, 5)}`;
    img.src = `/static/uploads/${encodeURIComponent(filename)}?${cacheBuster}`;
    img.alt = priceItem.title || '';
    img.dataset.originalFilename = priceItem.image_path; // для fallback
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
    const imageName = filename || 'Изображение';
    nameSpan.title = imageName;
    nameSpan.textContent = truncateText(imageName, 15);
    info.appendChild(nameSpan);

    // controls
    const controls = document.createElement('div');
    controls.className = 'image-controls';

    const cropBtn = document.createElement('button');
    cropBtn.className = 'crop-image';
    cropBtn.type = 'button';
    cropBtn.title = 'Настроить отображение';
    cropBtn.setAttribute('aria-label', 'Настроить кроппинг изображения');
    cropBtn.dataset.action = 'crop';
    cropBtn.dataset.priceId = priceItem.id;
    cropBtn.dataset.filename = priceItem.image_path;
    cropBtn.dataset.cropX = priceItem.crop_x || 50;
    cropBtn.dataset.cropY = priceItem.crop_y || 50;
    cropBtn.dataset.cropScale = priceItem.crop_scale || 1;
    cropBtn.textContent = '✂️';
    cropBtn.onclick = () => openPriceCropEditor(
        priceItem.id,
        priceItem.image_path,
        priceItem.crop_x || 50,
        priceItem.crop_y || 50,
        priceItem.crop_scale || 1.0
    );

    const delBtn = document.createElement('button');
    delBtn.className = 'delete-image';
    delBtn.type = 'button';
    delBtn.title = 'Удалить изображение';
    delBtn.setAttribute('aria-label', 'Удалить изображение');
    delBtn.dataset.action = 'delete';
    delBtn.dataset.priceId = priceItem.id;
    delBtn.textContent = '×';
    delBtn.onclick = () => deletePriceImage(priceItem.id);

    controls.appendChild(cropBtn);
    controls.appendChild(delBtn);

    item.appendChild(previewWrap);
    item.appendChild(info);
    item.appendChild(controls);

    return item;
}

// Обработка ошибки загрузки изображения (аналогично handleImageError)
function handlePriceImageError(img) {
    // Предотвращаем бесконечный цикл
    if (img.dataset.triedOriginal === 'true') {
        img.onerror = null;
        img.style.display = 'none';
        img.alt = 'Изображение не найдено';
        return;
    }

    // Пробуем загрузить оригинал как fallback
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

// Утилита для обрезки текста
function truncateText(text, max) {
    if (!text || text.length <= max) return text || '';
    return text.slice(0, max) + '…';
}

// ============== РЕДАКТОР КРОППИНГА ==============

let currentPriceCropEditor = null;

function openPriceCropEditor(priceId, imagePath, cropX, cropY, cropScale) {
    // Используем общий редактор кроппинга из crop-editor-api.js
    if (typeof openCropEditor !== 'function') {
        showAdminMessage('Редактор кроппинга не загружен', 'error');
        return;
    }

    const filename = imagePath.split(/[/\\]/).pop();

    // Сохраняем контекст
    currentPriceCropEditor = {
        priceId: priceId
    };

    // Открываем редактор кроппинга
    openCropEditor(priceId, filename, cropX, cropY, cropScale);

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
            const response = await fetch(`/admin/prices/${currentPriceCropEditor.priceId}/crop`, {
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

                    // Обновляем изображение в форме редактирования
                    setTimeout(() => {
                        const priceId = document.getElementById('editPriceId')?.value;
                        if (priceId) {
                            updatePriceImage(priceId);
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

// Обновление изображения позиции после кроппинга
async function updatePriceImage(priceId) {
    try {
        const response = await fetch(`/admin/prices/${priceId}`);
        const data = await response.json();

        if (response.ok && data.price_item) {
            displayPriceImage(data.price_item);
        }
    } catch (error) {
        console.error('Ошибка обновления изображения:', error);
    }
}

// ============== УДАЛЕНИЕ ИЗОБРАЖЕНИЯ ==============

async function deletePriceImage(priceId) {
    if (!confirm('Вы уверены, что хотите удалить изображение? Это действие нельзя отменить.')) {
        return;
    }

    try {
        const response = await fetch(`/admin/prices/${priceId}/image`, {
            method: 'DELETE'
        });

        const data = await response.json();

        if (data.success) {
            showAdminMessage('Изображение удалено', 'success');

            // Обновляем отображение в контейнере price_images
            const priceImagesContainer = document.getElementById('price_images');
            if (priceImagesContainer) {
                priceImagesContainer.innerHTML = `<p class="no-images">У позиции пока нет изображения</p>`;
            }

            // Очищаем превью загрузки если есть
            const fileInput = document.getElementById('editPriceImages');
            if (fileInput) {
                fileInput.value = '';
                const previewContainer = fileInput.parentNode.querySelector('.image-preview-container');
                if (previewContainer) {
                    previewContainer.remove();
                }
            }
        } else {
            showAdminMessage(data.error || 'Ошибка удаления изображения', 'error');
        }
    } catch (error) {
        console.error('Ошибка удаления изображения:', error);
        showAdminMessage('Ошибка удаления изображения', 'error');
    }
}

// ============== ЗАГРУЗКА ИЗОБРАЖЕНИЯ ==============

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
            showAdminMessage('Выберите файл для загрузки', 'error');
            return;
        }

        // Создаем FormData с изображением + данными из основной формы
        const formData = new FormData();
        formData.append('image', fileInput.files[0]);

        // Добавляем обязательные поля из формы редактирования
        const mainForm = document.getElementById('editPriceForm');
        if (mainForm) {
            formData.append('title', document.getElementById('editTitle')?.value || '');
            formData.append('description', document.getElementById('editDescription')?.value || '');
            formData.append('price_from', document.getElementById('editPriceFrom')?.value || '0');

            // Добавляем чекбоксы
            if (document.getElementById('editHasSpecifications')?.checked) {
                formData.append('has_specifications', 'on');
            }
            if (document.getElementById('editIsActive')?.checked) {
                formData.append('is_active', 'on');
            }

            // Добавляем характеристики если есть
            if (document.getElementById('editHasSpecifications')?.checked) {
                // Используем функцию из admin-prices.js
                if (typeof collectSpecifications === 'function') {
                    const specs = collectSpecifications('edit');
                    if (specs && specs.length > 0) {
                        formData.append('specifications', JSON.stringify(specs));
                    }
                }
            }
        }

        const spinner = document.getElementById('priceUploadSpinner');
        if (spinner) spinner.classList.remove('hidden');

        // Debug: логируем что отправляется
        console.log('=== Загрузка изображения для позиции прайса ===');
        for (let [key, value] of formData.entries()) {
            console.log(`${key}:`, value instanceof File ? `File(${value.name})` : value);
        }

        try {
            const response = await fetch(`/admin/prices/${priceId}/update`, {
                method: 'POST',
                body: formData
            });

            const data = await response.json();

            if (data.success) {
                showAdminMessage('Изображение загружено успешно', 'success');

                // Очищаем поле выбора файла
                fileInput.value = '';

                // Удаляем контейнер превью полностью
                const previewContainer = fileInput.parentNode.querySelector('.image-preview-container');
                if (previewContainer) {
                    previewContainer.remove();
                }

                // Обновляем отображение изображения
                await updatePriceImage(priceId);
            } else {
                showAdminMessage(data.error || 'Ошибка загрузки изображения', 'error');
            }
        } catch (error) {
            console.error('Ошибка загрузки изображения:', error);
            showAdminMessage('Ошибка загрузки изображения', 'error');
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

function formatFileSize(bytes) {
    if (!bytes) return '0 B';
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round((bytes / Math.pow(1024, i)) * 100) / 100 + ' ' + sizes[i];
}

// ============== ИНИЦИАЛИЗАЦИЯ ==============

// Экспортируем функции для использования в admin-prices.js
window.displayPriceImage = displayPriceImage;
window.openPriceCropEditor = openPriceCropEditor;
window.deletePriceImage = deletePriceImage;
window.updatePriceImage = updatePriceImage;
window.initPriceImageUploadForm = initPriceImageUploadForm;
window.initPriceImagePreview = initPriceImagePreview;

// Инициализируем при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    initPriceImageUploadForm();
    initPriceImagePreview();
});
