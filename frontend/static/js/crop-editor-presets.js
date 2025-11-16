// Предустановки и горячие клавиши для редактора кроппинга

// Предустановки для быстрого кроппинга
function applyCropPreset(presetName) {
    const presets = {
        center: { x: 50, y: 50, scale: 1 },
        topLeft: { x: 25, y: 25, scale: 1.2 },
        topRight: { x: 75, y: 25, scale: 1.2 },
        bottomLeft: { x: 25, y: 75, scale: 1.2 },
        bottomRight: { x: 75, y: 75, scale: 1.2 },
        zoomIn: { x: 50, y: 50, scale: 1.5 },
        zoomOut: { x: 50, y: 50, scale: 0.8 },
        fitWidth: { x: 50, y: 50, scale: 0.9 },
        fitHeight: { x: 50, y: 50, scale: 1.1 }
    };
    
    const preset = presets[presetName];
    if (preset) {
        // Используем анимацию для плавного перехода
        animateToPreset(preset);
        if (typeof showAdminMessage === 'function') {
            showAdminMessage(`Применена предустановка: ${getPresetDisplayName(presetName)}`, 'success');
        }
    }
}

// Получение отображаемого имени предустановки
function getPresetDisplayName(presetName) {
    const displayNames = {
        center: 'По центру',
        topLeft: 'Верхний левый угол',
        topRight: 'Верхний правый угол',
        bottomLeft: 'Нижний левый угол',
        bottomRight: 'Нижний правый угол',
        zoomIn: 'Увеличить',
        zoomOut: 'Уменьшить',
        fitWidth: 'По ширине',
        fitHeight: 'По высоте'
    };
    
    return displayNames[presetName] || presetName;
}

// Анимированный переход к предустановке
function animateToPreset(preset) {
    const duration = 500;
    
    // Анимируем каждый параметр
    animateSliderChange('cropX', preset.x, duration);
    animateSliderChange('cropY', preset.y, duration);
    animateSliderChange('cropScale', preset.scale, duration);
}

// Копирование настроек кроппинга в буфер
function copyCropSettings() {
    const settings = getCurrentCropValues();
    
    // Сохраняем в localStorage для последующего использования
    localStorage.setItem('cropSettings', JSON.stringify(settings));
    if (typeof showAdminMessage === 'function') {
        showAdminMessage('Настройки скопированы', 'success');
    }
}

// Вставка настроек кроппинга из буфера
function pasteCropSettings() {
    try {
        const savedSettings = localStorage.getItem('cropSettings');
        if (savedSettings) {
            const settings = JSON.parse(savedSettings);
            
            // Валидируем настройки
            if (isValidCropSettings(settings)) {
                setCropValues(settings.cropX, settings.cropY, settings.cropScale);
                if (typeof showAdminMessage === 'function') {
                    showAdminMessage('Настройки вставлены', 'success');
                }
            } else {
                if (typeof showAdminMessage === 'function') {
                    showAdminMessage('Некорректные сохраненные настройки', 'error');
                }
            }
        } else {
            if (typeof showAdminMessage === 'function') {
                showAdminMessage('Нет сохраненных настроек', 'error');
            }
        }
    } catch (error) {
        console.error('Ошибка при вставке настроек:', error);
        if (typeof showAdminMessage === 'function') {
            showAdminMessage('Ошибка при вставке настроек', 'error');
        }
    }
}

// Валидация настроек кроппинга
function isValidCropSettings(settings) {
    return (
        settings &&
        typeof settings.cropX === 'number' &&
        typeof settings.cropY === 'number' &&
        typeof settings.cropScale === 'number' &&
        settings.cropX >= 0 && settings.cropX <= 100 &&
        settings.cropY >= 0 && settings.cropY <= 100 &&
        settings.cropScale >= 0.5 && settings.cropScale <= 3
    );
}

// Горячие клавиши для редактора кроппинга
function initCropKeyboardShortcuts() {
    document.addEventListener('keydown', function(event) {
        // Проверяем, открыт ли редактор кроппинга
        if (!isCropEditorActive()) return;

        // НЕ блокируем горячие клавиши если фокус на input/textarea
        const activeElement = document.activeElement;
        if (activeElement && (activeElement.tagName === 'INPUT' || activeElement.tagName === 'TEXTAREA')) {
            return;
        }

        const step = event.shiftKey ? 10 : 1; // Больший шаг при зажатом Shift
        const scaleStep = event.shiftKey ? 0.2 : 0.1;
        
        switch (event.key) {
            case 'ArrowLeft':
                event.preventDefault();
                adjustSlider('cropX', -step);
                break;
            case 'ArrowRight':
                event.preventDefault();
                adjustSlider('cropX', step);
                break;
            case 'ArrowUp':
                event.preventDefault();
                adjustSlider('cropY', -step);
                break;
            case 'ArrowDown':
                event.preventDefault();
                adjustSlider('cropY', step);
                break;
            case '+':
            case '=':
                event.preventDefault();
                adjustSlider('cropScale', scaleStep);
                break;
            case '-':
                event.preventDefault();
                adjustSlider('cropScale', -scaleStep);
                break;
            case 'r':
                if (event.ctrlKey || event.metaKey) {
                    event.preventDefault();
                    resetCrop();
                }
                break;
            case 's':
                if (event.ctrlKey || event.metaKey) {
                    event.preventDefault();
                    saveCrop();
                }
                break;
            case 'c':
                if (event.ctrlKey || event.metaKey) {
                    event.preventDefault();
                    copyCropSettings();
                }
                break;
            case 'v':
                if (event.ctrlKey || event.metaKey) {
                    event.preventDefault();
                    pasteCropSettings();
                }
                break;
            case 'Escape':
                event.preventDefault();
                closeCropModal();
                break;
            case '1':
                event.preventDefault();
                applyCropPreset('center');
                break;
            case '2':
                event.preventDefault();
                applyCropPreset('zoomIn');
                break;
            case '3':
                event.preventDefault();
                applyCropPreset('zoomOut');
                break;
        }
    });
}

// Показ справки по горячим клавишам
function showKeyboardHelp() {
    const helpText = `🎯 Горячие клавиши редактора кроппинга:

🔄 Навигация:
• ← → ↑ ↓ - перемещение области кропа (+ Shift для больших шагов)
• + / - или = / - - масштабирование (+ Shift для больших шагов)

💾 Управление:
• Ctrl+S - сохранить настройки кропа
• Ctrl+R - сбросить к исходным настройкам
• Ctrl+C - копировать настройки кропа
• Ctrl+V - вставить настройки кропа

⚡ Предустановки:
• 1 - применить пресет "по центру"
• 2 - применить пресет "увеличить"
• 3 - применить пресет "уменьшить"`;

    alert(helpText);
}

// Экспорт функций
window.applyCropPreset = applyCropPreset;
window.copyCropSettings = copyCropSettings;
window.pasteCropSettings = pasteCropSettings;
window.initCropKeyboardShortcuts = initCropKeyboardShortcuts;
window.showKeyboardHelp = showKeyboardHelp;