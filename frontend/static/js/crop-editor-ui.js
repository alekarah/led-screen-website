// Управление интерфейсом редактора кроппинга

// Обновление отображаемых значений
function updateDisplayValues() {
    const cropX = document.getElementById('cropX')?.value || 50;
    const cropY = document.getElementById('cropY')?.value || 50;
    const cropScale = document.getElementById('cropScale')?.value || 1;
    
    const cropXValue = document.getElementById('cropXValue');
    const cropYValue = document.getElementById('cropYValue');
    const cropScaleValue = document.getElementById('cropScaleValue');
    
    if (cropXValue) cropXValue.textContent = Math.round(cropX);
    if (cropYValue) cropYValue.textContent = Math.round(cropY);
    if (cropScaleValue) cropScaleValue.textContent = parseFloat(cropScale).toFixed(1);
}

// Обновление превью при изменении слайдеров
function updateCropPreview() {
    const cropX = document.getElementById('cropX')?.value || 50;
    const cropY = document.getElementById('cropY')?.value || 50;
    const cropScale = document.getElementById('cropScale')?.value || 1;
    
    // Обновляем отображаемые значения
    updateDisplayValues();
    
    // Применяем стили к превью
    const previewImg = document.getElementById('cropPreviewImage');
    if (previewImg) {
        // Используем те же вычисления, что и в crop-editor-api.js
        const translateX = (cropX - 50) * 2; // Диапазон -100% до 100%
        const translateY = (cropY - 50) * 2; // Диапазон -100% до 100%
        
        previewImg.style.transform = `scale(${cropScale}) translate(${translateX}%, ${translateY}%)`;
        previewImg.style.transformOrigin = 'center center';
    }
}

// Инициализация слайдеров
function initCropSliders() {
    const sliders = ['cropX', 'cropY', 'cropScale'];
    
    sliders.forEach(sliderId => {
        const slider = document.getElementById(sliderId);
        if (slider) {
            // Обновляем превью при изменении слайдера
            slider.addEventListener('input', debounce(updateCropPreview, 50));
            
            // Также обновляем при отпускании мыши для точности
            slider.addEventListener('change', updateCropPreview);
        }
    });
}

// Изменение значения слайдера
function adjustSlider(sliderId, delta) {
    const slider = document.getElementById(sliderId);
    if (!slider) return;
    
    const currentValue = parseFloat(slider.value);
    const min = parseFloat(slider.min);
    const max = parseFloat(slider.max);
    const step = parseFloat(slider.step) || 1;
    
    let newValue = currentValue + delta;
    newValue = Math.max(min, Math.min(max, newValue));
    newValue = Math.round(newValue / step) * step;
    
    slider.value = newValue;
    updateCropPreview();
}

// Debounce функция для оптимизации
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Анимация изменения значений
function animateSliderChange(sliderId, targetValue, duration = 300) {
    const slider = document.getElementById(sliderId);
    if (!slider) return;
    
    const startValue = parseFloat(slider.value);
    const difference = targetValue - startValue;
    const startTime = performance.now();
    
    function animate(currentTime) {
        const elapsed = currentTime - startTime;
        const progress = Math.min(elapsed / duration, 1);
        
        // Easing function (ease-out)
        const easeOut = 1 - Math.pow(1 - progress, 3);
        
        const currentValue = startValue + (difference * easeOut);
        slider.value = currentValue;
        updateCropPreview();
        
        if (progress < 1) {
            requestAnimationFrame(animate);
        }
    }
    
    requestAnimationFrame(animate);
}

// Инициализация UI компонентов
function initCropUI() {
    // Инициализируем слайдеры
    initCropSliders();
    
    // Добавляем обработчики для кнопок
    const resetBtn = document.querySelector('[onclick="resetCrop()"]');
    const saveBtn = document.querySelector('[onclick="saveCrop()"]');
    
    if (resetBtn) {
        resetBtn.addEventListener('click', (e) => {
            e.preventDefault();
            resetCrop();
        });
    }
    
    if (saveBtn) {
        saveBtn.addEventListener('click', (e) => {
            e.preventDefault();
            saveCrop();
        });
    }
}

// Показ индикатора загрузки в превью
function showPreviewLoading() {
    const previewImg = document.getElementById('cropPreviewImage');
    if (previewImg) {
        previewImg.style.opacity = '0.5';
        previewImg.style.filter = 'blur(1px)';
    }
}

// Скрытие индикатора загрузки в превью
function hidePreviewLoading() {
    const previewImg = document.getElementById('cropPreviewImage');
    if (previewImg) {
        previewImg.style.opacity = '1';
        previewImg.style.filter = 'none';
    }
}

// Экспорт функций
window.updateDisplayValues = updateDisplayValues;
window.updateCropPreview = updateCropPreview;
window.initCropSliders = initCropSliders;
window.adjustSlider = adjustSlider;
window.animateSliderChange = animateSliderChange;
window.initCropUI = initCropUI;
window.showPreviewLoading = showPreviewLoading;
window.hidePreviewLoading = hidePreviewLoading;