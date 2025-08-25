// Базовые функции для админки

// Показ сообщений
function showAdminMessage(message, type = 'success', timeout = 4000) {
    let root = document.getElementById('message');
    if (!root) {
        root = document.createElement('div');
        root.id = 'message';
        document.body.appendChild(root);
    }

    // Антидубликатор: если тот же текст прилетел за последние 1500ms — не показываем второй раз
    const now = Date.now();
    window.__lastToast = window.__lastToast || { text: '', ts: 0 };
    if (window.__lastToast.text === message && (now - window.__lastToast.ts) < 1500) {
        return;
    }
    window.__lastToast = { text: message, ts: now };

    const el = document.createElement('div');
    el.className = 'toast ' + (type === 'error' ? 'error' : 'success');
    el.innerHTML = `
        <span>${message}</span>
        <button class="close" aria-label="Закрыть">&times;</button>
    `;

    el.querySelector('.close').addEventListener('click', () => {
        root.removeChild(el);
    });

    root.appendChild(el);

    if (timeout) {
        setTimeout(() => {
            el.style.transition = 'opacity .2s ease';
            el.style.opacity = '0';
            setTimeout(() => root.contains(el) && root.removeChild(el), 200);
        }, timeout);
    }
}

// Обработка ошибок fetch запросов
async function handleFetchResponse(response) {
    const result = await response.json();
    
    if (response.ok) {
        if (result.message) {
            showAdminMessage(result.message, 'success');
        }
        return result;
    } else {
        const errorMessage = result.error || `Ошибка сервера: ${response.status}`;
        showAdminMessage(errorMessage, 'error');
        throw new Error(errorMessage);
    }
}

// Универсальная функция для отправки форм
async function submitForm(formElement, url, method = 'POST') {
    try {
        const formData = new FormData(formElement);
        
        const response = await fetch(url, {
            method: method,
            body: formData
        });
        
        return await handleFetchResponse(response);
    } catch (error) {
        console.error('Ошибка отправки формы:', error);
        showAdminMessage('Ошибка: ' + error.message, 'error');
        throw error;
    }
}

// Универсальная функция для GET запросов
async function fetchData(url) {
    try {
        const response = await fetch(url);
        return await handleFetchResponse(response);
    } catch (error) {
        console.error('Ошибка загрузки данных:', error);
        showAdminMessage('Ошибка загрузки: ' + error.message, 'error');
        throw error;
    }
}

// Универсальная функция для DELETE запросов
async function deleteData(url) {
    try {
        const response = await fetch(url, {
            method: 'DELETE'
        });
        
        return await handleFetchResponse(response);
    } catch (error) {
        console.error('Ошибка удаления:', error);
        showAdminMessage('Ошибка удаления: ' + error.message, 'error');
        throw error;
    }
}

// Подтверждение действий
function confirmAction(message, callback) {
    if (confirm(message)) {
        callback();
    }
}

// Закрытие модальных окон по клику вне их
function setupModalCloseHandlers() {
    window.addEventListener('click', function(event) {
        // Находим все модальные окна
        const modals = document.querySelectorAll('.modal');
        
        modals.forEach(modal => {
            if (event.target === modal) {
                modal.style.display = 'none';
            }
        });
    });
    
    // Закрытие по клавише Escape
    document.addEventListener('keydown', function(event) {
        if (event.key === 'Escape') {
            const visibleModals = document.querySelectorAll('.modal[style*="block"]');
            visibleModals.forEach(modal => {
                modal.style.display = 'none';
            });
        }
    });
}

// Функции для работы с модальными окнами
function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.style.display = 'block';
        document.body.style.overflow = 'hidden'; // Блокируем прокрутку фона
    }
}

function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.style.display = 'none';
        document.body.style.overflow = 'auto'; // Возвращаем прокрутку
    }
}

// Утилиты для работы с формами
function resetForm(formId) {
    const form = document.getElementById(formId);
    if (form) {
        form.reset();
    }
}

function clearValidationErrors() {
    const errorElements = document.querySelectorAll('.field-error');
    errorElements.forEach(el => el.remove());
    
    const invalidFields = document.querySelectorAll('.invalid');
    invalidFields.forEach(field => field.classList.remove('invalid'));
}

// Валидация форм
function validateRequired(formElement) {
    const requiredFields = formElement.querySelectorAll('[required]');
    let isValid = true;
    
    clearValidationErrors();
    
    requiredFields.forEach(field => {
        if (!field.value.trim()) {
            isValid = false;
            field.classList.add('invalid');
            
            const error = document.createElement('div');
            error.className = 'field-error';
            error.textContent = 'Это поле обязательно для заполнения';
            error.style.color = '#dc3545';
            error.style.fontSize = '0.875rem';
            error.style.marginTop = '0.25rem';
            
            field.parentNode.appendChild(error);
        }
    });
    
    return isValid;
}

// Утилиты для работы с изображениями
function createImagePreview(file, container) {
    const reader = new FileReader();
    
    reader.onload = function(e) {
        const img = document.createElement('img');
        img.src = e.target.result;
        img.style.maxWidth = '100px';
        img.style.maxHeight = '100px';
        img.style.objectFit = 'cover';
        img.style.borderRadius = '4px';
        img.style.margin = '0.25rem';
        
        container.appendChild(img);
    };
    
    reader.readAsDataURL(file);
}

// Форматирование данных
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

// Дебаунс функция для оптимизации
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

// Инициализация базовых функций
document.addEventListener('DOMContentLoaded', function() {
    // Настраиваем обработчики закрытия модальных окон
    setupModalCloseHandlers();
    
    // Добавляем общие обработчики для всех форм
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        // Очищаем ошибки валидации при изменении полей
        const inputs = form.querySelectorAll('input, textarea, select');
        inputs.forEach(input => {
            input.addEventListener('input', clearValidationErrors);
        });
    });
});

// ЭКСПОРТ ФУНКЦИЙ ПОСЛЕ DOMContentLoaded
window.showAdminMessage = showAdminMessage;
window.handleFetchResponse = handleFetchResponse;
window.submitForm = submitForm;
window.fetchData = fetchData;
window.deleteData = deleteData;