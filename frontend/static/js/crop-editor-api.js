// API для сохранения настроек кроппинга

// Глобальная переменная для хранения ID текущего редактируемого изображения
let currentEditingImageId = null;

// Открытие редактора кроппинга
function openCropEditor(imageId, filename, cropX = 50, cropY = 50, cropScale = 1) {
    currentEditingImageId = imageId;
    
    // Получаем элементы
    const modal = document.getElementById('cropModal');
    const previewImg = document.getElementById('cropPreviewImage');
    
    if (!modal || !previewImg) {
        console.error('Элементы редактора кроппинга не найдены');
        return;
    }
    
    // Устанавливаем изображение
    previewImg.src = `/static/uploads/${encodeURIComponent(filename)}`;
    previewImg.alt = filename;
    
    // Ждем загрузки изображения и устанавливаем значения
    previewImg.onload = () => {
        // СНАЧАЛА центрируем изображение принудительно
        previewImg.style.transform = 'scale(1) translate(0%, 0%)';
        previewImg.style.transformOrigin = 'center center';
        
        // Получаем ползунки для принудительного центрирования
        const cropXSlider = document.getElementById('cropX');
        const cropYSlider = document.getElementById('cropY');
        const cropScaleSlider = document.getElementById('cropScale');
        
        // Устанавливаем центральные значения ПЕРЕД вызовом setCropValues
        if (cropXSlider) cropXSlider.value = 50;
        if (cropYSlider) cropYSlider.value = 50;
        if (cropScaleSlider) cropScaleSlider.value = 1;

        // Обновляем отображение с центральными значениями
        updatePreviewTransform();

        // ПОТОМ устанавливаем переданные значения (если нужно)
        if (cropX !== 50 || cropY !== 50 || cropScale !== 1) {
            setCropValues(cropX, cropY, cropScale);
        }
        
        // ПРИНУДИТЕЛЬНО добавляем обработчики событий для слайдеров
        // Удаляем старые обработчики
        if (cropXSlider) {
            cropXSlider.removeEventListener('input', updatePreviewTransform);
            cropXSlider.addEventListener('input', updatePreviewTransform);
        }
        if (cropYSlider) {
            cropYSlider.removeEventListener('input', updatePreviewTransform);
            cropYSlider.addEventListener('input', updatePreviewTransform);
        }
        if (cropScaleSlider) {
            cropScaleSlider.removeEventListener('input', updatePreviewTransform);
            cropScaleSlider.addEventListener('input', updatePreviewTransform);
        }
        
        // ПРИНУДИТЕЛЬНО добавляем обработчики для кнопок
        const resetBtn = document.querySelector('button[onclick="resetCrop()"]');
        const saveBtn = document.querySelector('button[onclick="saveCrop()"]');
        
        if (resetBtn) {
            resetBtn.removeAttribute('onclick');
            resetBtn.addEventListener('click', (e) => {
                e.preventDefault();
                resetCrop();
            });
        }
        
        if (saveBtn) {
            saveBtn.removeAttribute('onclick');
            saveBtn.addEventListener('click', (e) => {
                e.preventDefault();
                saveCrop();
            });
        }
    };
    
    // Показываем модальное окно
    modal.style.display = 'block';
    document.body.style.overflow = 'hidden';
}

// Обновление превью изображения
function updatePreviewTransform() {
    const previewImg = document.getElementById('cropPreviewImage');
    if (!previewImg) return;
    
    const cropXSlider = document.getElementById('cropX');
    const cropYSlider = document.getElementById('cropY');
    const cropScaleSlider = document.getElementById('cropScale');
    
    if (!cropXSlider || !cropYSlider || !cropScaleSlider) return;
    
    const cropX = parseFloat(cropXSlider.value);
    const cropY = parseFloat(cropYSlider.value);
    const cropScale = parseFloat(cropScaleSlider.value);

    const cropXValue = document.getElementById('cropXValue');
    const cropYValue = document.getElementById('cropYValue');
    const cropScaleValue = document.getElementById('cropScaleValue');
    
    if (cropXValue) cropXValue.textContent = Math.round(cropX);
    if (cropYValue) cropYValue.textContent = Math.round(cropY);
    if (cropScaleValue) cropScaleValue.textContent = cropScale.toFixed(1);    
    
    // Преобразуем значения в CSS translate
    // cropX=0% показывает правый край, cropX=100% показывает левый край
    const translateX = (cropX - 50) * 2; // Диапазон -100% до 100%
    const translateY = (cropY - 50) * 2; // Диапазон -100% до 100%
    
    // Применяем трансформацию
    previewImg.style.transform = `scale(${cropScale}) translate(${translateX}%, ${translateY}%)`;
    previewImg.style.transformOrigin = 'center center';
}

// Сохранение настроек кроппинга
async function saveCrop() {
    if (!currentEditingImageId) {
        showCropMessage('Ошибка: не выбрано изображение для редактирования', 'error');
        return;
    }
    
    try {
        // Получаем текущие значения
        const cropX = parseFloat(document.getElementById('cropX').value);
        const cropY = parseFloat(document.getElementById('cropY').value);
        const cropScale = parseFloat(document.getElementById('cropScale').value);
        
        // Преобразуем значения в правильный диапазон для сервера
        const serverCropX = cropX; // Диапазон 0-100
        const serverCropY = cropY; // Диапазон 0-100
        
        // Отправляем запрос на сервер
        const response = await fetch(`/admin/images/${currentEditingImageId}/crop`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                image_id: parseInt(currentEditingImageId),
                crop_x: serverCropX,
                crop_y: serverCropY,
                crop_scale: cropScale
            })
        });
        
        const result = await response.json();
        
        if (response.ok) {
            showCropMessage('Настройки изображения сохранены!', 'success');
            
            // Закрываем модальное окно
            setTimeout(() => {
                closeCropModal();
                
                // Обновляем изображения в основной форме
                const projectId = document.getElementById('edit_project_id')?.value;
                if (projectId) {
                    updateProjectImages(projectId);
                }
            }, 1000);
            
        } else {
            throw new Error(result.error || 'Ошибка сохранения');
        }
        
    } catch (error) {
        console.error('Ошибка сохранения кроппинга:', error);
        showCropMessage('Ошибка при сохранении: ' + error.message, 'error');
    }
}

// Обновление изображений проекта после кроппинга
async function updateProjectImages(projectId) {
    try {
        const response = await fetch(`/admin/projects/${projectId}`);
        const data = await response.json();
        
        if (response.ok && data.project && data.project.images) {
            // Обновляем изображения в DOM
            if (typeof fillProjectImages === 'function') {
                fillProjectImages(data.project.images);
            }
        }
    } catch (error) {
        console.error('Ошибка обновления изображений:', error);
    }
}

// Функция показа сообщений в редакторе кроппинга
function showCropMessage(message, type = 'success') {
    // Пытаемся использовать глобальную функцию showAdminMessage
    if (typeof window.showAdminMessage === 'function') {
        window.showAdminMessage(message, type);
        return;
    }
    
    // Если нет - показываем обычный alert
    if (type === 'error') {
        alert('❌ ' + message);
    } else {
        alert('✅ ' + message);
    }
}

// Экспорт функций
window.openCropEditor = openCropEditor;
window.updatePreviewTransform = updatePreviewTransform;
window.saveCrop = saveCrop;
window.updateProjectImages = updateProjectImages;
window.showCropMessage = showCropMessage;