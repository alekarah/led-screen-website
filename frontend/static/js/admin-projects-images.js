// Работа с изображениями проектов

// Заполнение изображений проекта
function fillProjectImages(images) {
    const imagesDiv = document.getElementById('project_images');
    if (!imagesDiv) {
        console.warn('Контейнер изображений не найден');
        return;
    }
    
    imagesDiv.innerHTML = '';
    
    if (!images || !Array.isArray(images)) {
        imagesDiv.innerHTML = '<p class="no-images">Изображения не загружены</p>';
        return;
    }
    
    if (images.length === 0) {
        imagesDiv.innerHTML = '<p class="no-images">У проекта пока нет изображений</p>';
        return;
    }
    
    images.forEach(image => {
        const imageHTML = createImageHTML(image);
        imagesDiv.innerHTML += imageHTML;
    });
    
    // Добавляем обработчики событий для новых элементов
    initImageEventHandlers();
}

// Создание HTML для изображения
function createImageHTML(image) {
    const cropX = image.crop_x !== undefined ? image.crop_x : 50;
    const cropY = image.crop_y !== undefined ? image.crop_y : 50;
    const cropScale = image.crop_scale || 1;
    
    // Преобразуем для CSS отображения
    const translateX = (cropX - 50) * 2;
    const translateY = (cropY - 50) * 2;
    
    return `
        <div class="image-item" data-image-id="${image.id}">
            <div class="image-preview-container">
                <img src="/static/uploads/${encodeURIComponent(image.filename)}" 
                     alt="${escapeHtml(image.alt || image.original_name || '')}" 
                     onerror="handleImageError(this)"
                     style="transform: scale(${cropScale}) translate(${translateX}%, ${translateY}%); object-position: center center; transform-origin: center center;">
                
                <!-- Индикатор загрузки -->
                <div class="image-loading hidden">
                    <div class="loading-spinner"></div>
                </div>
            </div>
            
            <!-- Информация об изображении -->
            <div class="image-info">
                <span class="image-name" title="${escapeHtml(image.original_name || image.filename)}">
                    ${truncateText(image.original_name || image.filename, 15)}
                </span>
                <span class="image-size">${formatFileSize(image.file_size || 0)}</span>
            </div>
            
            <!-- Кнопки управления -->
            <div class="image-controls">
                <button class="crop-image" 
                        onclick="openCropEditor(${image.id}, '${escapeHtml(image.filename)}', ${cropX}, ${cropY}, ${cropScale})" 
                        title="Настроить отображение"
                        aria-label="Настроить кроппинг изображения">
                    ✂️
                </button>
                <button class="delete-image" 
                        onclick="deleteImage(${image.id})" 
                        title="Удалить изображение"
                        aria-label="Удалить изображение">
                    &times;
                </button>
            </div>
        </div>
    `;
}

// Удаление изображения
async function deleteImage(imageId) {
    if (!imageId) {
        showAdminMessage('Некорректный ID изображения', 'error');
        return;
    }
    
    confirmAction('Удалить это изображение?', async () => {
        // Показываем индикатор загрузки для конкретного изображения
        showImageLoading(imageId, true);
        
        try {
            await deleteData(`/admin/images/${imageId}`);
            
            showAdminMessage('Изображение удалено', 'success');
            
            // Удаляем элемент из DOM с анимацией
            removeImageFromDOM(imageId);
            
            // Перезагружаем изображения проекта для синхронизации
            setTimeout(async () => {
                const projectId = document.getElementById('edit_project_id')?.value;
                if (projectId) {
                    try {
                        const data = await fetchData(`/admin/projects/${projectId}`);
                        fillProjectImages(data.project.images || []);
                    } catch (error) {
                        console.error('Ошибка обновления списка изображений:', error);
                    }
                }
            }, 500);
            
        } catch (error) {
            console.error('Ошибка удаления изображения:', error);
            showAdminMessage('Ошибка при удалении изображения: ' + error.message, 'error');
        } finally {
            showImageLoading(imageId, false);
        }
    });
}

// Удаление изображения из DOM с анимацией
function removeImageFromDOM(imageId) {
    const imageElement = document.querySelector(`[data-image-id="${imageId}"]`);
    if (imageElement) {
        imageElement.style.animation = 'fadeOut 0.3s ease-out';
        setTimeout(() => {
            imageElement.remove();
        }, 300);
    }
}

// Предварительный просмотр загружаемых изображений
function initImagePreview() {
    const imageInputs = document.querySelectorAll('input[type="file"][accept*="image"]');
    
    imageInputs.forEach(input => {
        input.addEventListener('change', function(e) {
            const files = e.target.files;
            const previewContainer = getOrCreatePreviewContainer(input);
            
            // Очищаем предыдущие превью
            previewContainer.innerHTML = '';
            
            if (files.length === 0) return;
            
            Array.from(files).forEach((file, index) => {
                if (file.type.startsWith('image/')) {
                    createImagePreview(file, previewContainer, index);
                }
            });
        });
    });
}

// Создание превью для загружаемого изображения
function createImagePreview(file, container, index) {
    const reader = new FileReader();
    
    reader.onload = function(e) {
        const previewHTML = `
            <div class="upload-preview" data-index="${index}">
                <img src="${e.target.result}" alt="Превью" style="width: 80px; height: 60px; object-fit: cover; border-radius: 4px;">
                <div class="preview-info">
                    <span class="preview-name">${truncateText(file.name, 12)}</span>
                    <span class="preview-size">${formatFileSize(file.size)}</span>
                </div>
                <button type="button" class="remove-preview" onclick="removePreview(this)" title="Удалить из списка">&times;</button>
            </div>
        `;
        
        container.innerHTML += previewHTML;
    };
    
    reader.readAsDataURL(file);
}

// Удаление превью из списка
function removePreview(button) {
    const previewElement = button.closest('.upload-preview');
    if (previewElement) {
        const index = parseInt(previewElement.dataset.index);
        const input = previewElement.closest('.form-group').querySelector('input[type="file"]');
        
        // Удаляем файл из списка
        if (input && input.files) {
            const dt = new DataTransfer();
            Array.from(input.files).forEach((file, i) => {
                if (i !== index) {
                    dt.items.add(file);
                }
            });
            input.files = dt.files;
        }
        
        previewElement.remove();
    }
}

// Получение или создание контейнера для превью
function getOrCreatePreviewContainer(input) {
    let container = input.parentNode.querySelector('.image-preview-container');
    
    if (!container) {
        container = document.createElement('div');
        container.className = 'image-preview-container';
        container.style.cssText = `
            margin-top: 1rem;
            display: flex;
            flex-wrap: wrap;
            gap: 0.5rem;
            max-height: 200px;
            overflow-y: auto;
            padding: 0.5rem;
            border: 1px dashed #ddd;
            border-radius: 4px;
        `;
        
        input.parentNode.appendChild(container);
    }
    
    return container;
}

// Обработка ошибок загрузки изображений
function handleImageError(img) {
    img.src = '/static/images/placeholder.jpg';
    img.alt = 'Изображение не найдено';
    img.style.opacity = '0.5';
}

// Показ/скрытие индикатора загрузки для изображения
function showImageLoading(imageId, show) {
    const imageElement = document.querySelector(`[data-image-id="${imageId}"]`);
    if (!imageElement) return;
    
    const loading = imageElement.querySelector('.image-loading');
    const controls = imageElement.querySelector('.image-controls');
    
    if (show) {
        if (loading) loading.classList.remove('hidden');
        if (controls) controls.style.opacity = '0.5';
    } else {
        if (loading) loading.classList.add('hidden');
        if (controls) controls.style.opacity = '1';
    }
}

// Инициализация обработчиков событий для изображений
function initImageEventHandlers() {
    // Добавляем hover эффекты
    const imageItems = document.querySelectorAll('.image-item');
    imageItems.forEach(item => {
        item.addEventListener('mouseenter', function() {
            this.style.transform = 'scale(1.02)';
        });
        
        item.addEventListener('mouseleave', function() {
            this.style.transform = 'scale(1)';
        });
    });
}

// Утилитарные функции
function truncateText(text, maxLength) {
    if (!text || text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
}

function formatFileSize(bytes) {
    if (!bytes) return '0 B';
    
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
}

function escapeHtml(text) {
    if (!text) return '';
    
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Экспорт функций
window.fillProjectImages = fillProjectImages;
window.deleteImage = deleteImage;
window.initImagePreview = initImagePreview;
window.handleImageError = handleImageError;
window.removePreview = removePreview;