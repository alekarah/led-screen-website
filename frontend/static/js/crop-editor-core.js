// Основная логика редактора кроппинга

// Текущее редактируемое изображение
let currentImageId = null;
let originalImageSrc = '';

// Установка значений в слайдеры
function setCropValues(cropX, cropY, cropScale) {
    const cropXSlider = document.getElementById('cropX');
    const cropYSlider = document.getElementById('cropY');
    const cropScaleSlider = document.getElementById('cropScale');
    
    if (cropXSlider) cropXSlider.value = cropX;
    if (cropYSlider) cropYSlider.value = cropY;
    if (cropScaleSlider) cropScaleSlider.value = cropScale;
    
    // Обновляем отображаемые значения
    updateDisplayValues();
    
    // Обновляем превью
    updatePreviewTransform();
}

// Сброс настроек к значениям по умолчанию
function resetCrop() {
    setCropValues(50, 50, 1);
    showCropMessage('Настройки сброшены к значениям по умолчанию', 'success');
}

// Закрытие модального окна редактора кроппинга
function closeCropModal() {
    const modal = document.getElementById('cropModal');
    if (modal) {
        modal.style.display = 'none';
        document.body.style.overflow = 'auto';
    }
    
    currentImageId = null;
    originalImageSrc = '';
    
    // Сбрасываем превью
    const previewImg = document.getElementById('cropPreviewImage');
    if (previewImg) {
        previewImg.src = '';
        previewImg.style.transform = '';
    }
}

// Получение текущих значений кроппинга
function getCurrentCropValues() {
    const cropX = document.getElementById('cropX')?.value || 50;
    const cropY = document.getElementById('cropY')?.value || 50;
    const cropScale = document.getElementById('cropScale')?.value || 1;
    
    return {
        cropX: parseFloat(cropX),
        cropY: parseFloat(cropY),
        cropScale: parseFloat(cropScale)
    };
}

// Проверка активности редактора
function isCropEditorActive() {
    const cropModal = document.getElementById('cropModal');
    return cropModal && cropModal.style.display !== 'none';
}

// Добавляем функцию showCropMessage для использования в resetCrop
function showCropMessage(message, type = 'success') {
    if (typeof window.showAdminMessage === 'function') {
        window.showAdminMessage(message, type);
        return;
    }
    
    if (type === 'error') {
        alert('❌ ' + message);
    } else {
        alert('✅ ' + message);
    }
}

// Экспорт функций
window.setCropValues = setCropValues;
window.resetCrop = resetCrop;
window.closeCropModal = closeCropModal;
window.getCurrentCropValues = getCurrentCropValues;
window.isCropEditorActive = isCropEditorActive;