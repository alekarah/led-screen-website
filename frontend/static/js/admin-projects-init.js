// Инициализация и утилиты для управления проектами

// Инициализация всех функций при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    try {
        initializeModules();
        initCommonEventHandlers();
        setupGlobalSettings();
        focusProjectFromQuery();
    } catch (error) {
        console.error('❌ Ошибка инициализации модулей:', error);
        showAdminMessage('Ошибка загрузки модулей управления проектами', 'error');
    }
});

// Инициализация модулей
function initializeModules() {
    if (typeof initProjectCreation === 'function') initProjectCreation();
    if (typeof initProjectEditing === 'function') initProjectEditing();
    if (typeof initImagePreview === 'function') initImagePreview();
    if (typeof initProjectSorting === 'function' && window.Sortable) initProjectSorting();
}

// Инициализация общих обработчиков событий
function initCommonEventHandlers() {
    // Обработчик для закрытия модальных окон по ESC
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape') {
            closeAllModals();
        }
    });
    
    // Обработчик для закрытия модальных окон по клику по оверлею
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', function (e) {
            if (e.target === modal) { // клик по подложке
                closeModal(modal.id);
            }
        });
    });
    
    // Обработчик для автосохранения форм
    initAutoSave();
    
    // Обработчик для drag & drop файлов
    initDragAndDrop();

    // Кнопка "Создать проект"
    const btn = document.getElementById('openCreateProjectModal');
    if (btn) btn.onclick = openCreateProjectModal;
}

// Настройка глобальных параметров
function setupGlobalSettings() {
    // Настройки для AJAX запросов
    setupAjaxDefaults();
    
    // Настройки для обработки ошибок
    setupErrorHandling();
    
    // Настройки производительности
    setupPerformanceOptimizations();
}

// Закрытие всех модальных окон
function closeAllModals() {
    const modals = document.querySelectorAll('.modal');
    modals.forEach(modal => {
        if (modal.style.display !== 'none') {
            closeModal(modal.id);
        }
    });
}

// Инициализация автосохранения форм
function initAutoSave() {
    const forms = document.querySelectorAll('form[data-autosave]');
    
    forms.forEach(form => {
        const inputs = form.querySelectorAll('input, textarea, select');
        
        inputs.forEach(input => {
            input.addEventListener('input', debounce(() => {
                saveFormData(form);
            }, 1000));
        });
    });
}

// Сохранение данных формы в localStorage
function saveFormData(form) {
    const formData = new FormData(form);
    const data = {};
    
    for (let [key, value] of formData.entries()) {
        data[key] = value;
    }
    
    const formId = form.id || 'unknown_form';
    localStorage.setItem(`autosave_${formId}`, JSON.stringify(data));
}

// Восстановление данных формы из localStorage
function restoreFormData(formId) {
    const savedData = localStorage.getItem(`autosave_${formId}`);
    if (!savedData) return;
    
    try {
        const data = JSON.parse(savedData);
        const form = document.getElementById(formId);
        if (!form) return;
        
        Object.entries(data).forEach(([key, value]) => {
            const input = form.querySelector(`[name="${key}"]`);
            if (input) {
                if (input.type === 'checkbox') {
                    input.checked = value === 'on';
                } else {
                    input.value = value;
                }
            }
        });
    } catch (error) {
        console.error('Ошибка восстановления данных формы:', error);
    }
}

// Инициализация drag & drop для файлов
function initDragAndDrop() {
    const fileInputs = document.querySelectorAll('input[type="file"]');
    
    fileInputs.forEach(input => {
        const container = input.closest('.form-group');
        if (!container) return;
        
        // Добавляем визуальные эффекты для drag & drop
        container.addEventListener('dragover', function(e) {
            e.preventDefault();
            this.classList.add('drag-over');
        });
        
        container.addEventListener('dragleave', function(e) {
            e.preventDefault();
            this.classList.remove('drag-over');
        });
        
        container.addEventListener('drop', function(e) {
            e.preventDefault();
            this.classList.remove('drag-over');
            
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                input.files = files;
                input.dispatchEvent(new Event('change'));
            }
        });
    });
}

// Настройки AJAX запросов
function setupAjaxDefaults() {
    // Можно добавить общие настройки для fetch запросов
    window.defaultFetchOptions = {
        credentials: 'same-origin',
        headers: {
            'X-Requested-With': 'XMLHttpRequest'
        }
    };
}

// Настройка обработки ошибок
function setupErrorHandling() {
    // Глобальный обработчик ошибок
    window.addEventListener('error', function(e) {
        console.error('Глобальная ошибка:', e.error);
    });
    
    // Обработчик необработанных промисов
    window.addEventListener('unhandledrejection', function(e) {
        console.error('Необработанная ошибка промиса:', e.reason);
    });
}

// Оптимизации производительности
function setupPerformanceOptimizations() {
    // Debounce функция для оптимизации событий
    window.debounce = function(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    };
    
    // Throttle функция для ограничения частоты вызовов
    window.throttle = function(func, limit) {
        let inThrottle;
        return function() {
            const args = arguments;
            const context = this;
            if (!inThrottle) {
                func.apply(context, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    };
}

// Утилитарные функции
function generateUniqueId() {
    return 'id_' + Math.random().toString(36).substr(2, 9) + '_' + Date.now();
}

function formatDateTime(date) {
    if (!date) return '';
    
    const d = new Date(date);
    return d.toLocaleString('ru-RU', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function copyToClipboard(text) {
    if (navigator.clipboard) {
        navigator.clipboard.writeText(text).then(() => {
            showAdminMessage('Скопировано в буфер обмена', 'success');
        });
    } else {
        // Fallback для старых браузеров
        const textArea = document.createElement('textarea');
        textArea.value = text;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
        showAdminMessage('Скопировано в буфер обмена', 'success');
    }
}

// Фокус на проект по параметру URL (?focus_id=123)
function focusProjectFromQuery() {
    const params = new URLSearchParams(window.location.search);
    const id = params.get('focus_id');
    if (!id) return;

    const el = document.querySelector(`.project-item[data-project-id="${id}"]`);
    if (!el) return;

    try {
        el.scrollIntoView({ behavior: 'smooth', block: 'center' });
    } catch (_) {
        el.scrollIntoView();
    }

    el.classList.add('project-focus');
    setTimeout(() => el.classList.remove('project-focus'), 2500);
}

// Диагностические функции
function checkModulesStatus() {
    const modules = [
        'initProjectCreation',
        'initImagePreview'
    ];
    
    const status = {};
    modules.forEach(module => {
        status[module] = typeof window[module] === 'function';
    });
    
    console.table(status);
    return status;
}

function getProjectsStats() {
    const projectItems = document.querySelectorAll('.project-item');
    const imageItems = document.querySelectorAll('.image-item');
    
    return {
        totalProjects: projectItems.length,
        totalImages: imageItems.length,
        featuredProjects: document.querySelectorAll('.project-item [title*="⭐"]').length
    };
}

// Экспорт утилитарных функций
window.generateUniqueId = generateUniqueId;
window.formatDateTime = formatDateTime;
window.copyToClipboard = copyToClipboard;
window.checkModulesStatus = checkModulesStatus;
window.getProjectsStats = getProjectsStats;
window.restoreFormData = restoreFormData;

// Экспорт для отладки
window.AdminProjectsDebug = {
    checkModulesStatus,
    getProjectsStats,
    closeAllModals
};