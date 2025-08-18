// Форма обратной связи на странице контактов

// Инициализация формы контактов
document.addEventListener('DOMContentLoaded', function() {
    initContactForm();
    initPhoneMask();
    initFormValidation();
});

// Инициализация основной формы
function initContactForm() {
    const contactForm = document.getElementById('contactForm');
    if (!contactForm) return;

    contactForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        // Валидация формы
        if (!validateContactForm(contactForm)) {
            return;
        }
        
        // Показываем состояние отправки
        setSubmitState(true);
        
        try {
            const formData = new FormData(contactForm);
            const data = Object.fromEntries(formData.entries());
            
            const result = await submitContactForm(data);
            
            if (result.success) {
                showSuccessMessage();
                resetContactForm();
            } else {
                showErrorMessage(result.message || 'Ошибка отправки формы');
            }
        } catch (error) {
            console.error('Ошибка отправки формы:', error);
            showErrorMessage('Произошла ошибка. Попробуйте позже или свяжитесь по телефону.');
        } finally {
            setSubmitState(false);
        }
    });
}

// Валидация формы контактов
function validateContactForm(form) {
    const errors = [];
    
    // Проверяем обязательные поля
    const name = form.querySelector('#name').value.trim();
    const phone = form.querySelector('#phone').value.trim();
    const privacy = form.querySelector('#privacy').checked;
    
    if (!name) {
        errors.push('Укажите ваше имя');
        highlightField('name');
    } else if (name.length < 2) {
        errors.push('Имя должно содержать минимум 2 символа');
        highlightField('name');
    }
    
    if (!phone) {
        errors.push('Укажите номер телефона');
        highlightField('phone');
    } else if (!isValidPhone(phone)) {
        errors.push('Укажите корректный номер телефона');
        highlightField('phone');
    }
    
    if (!privacy) {
        errors.push('Необходимо согласие на обработку персональных данных');
        highlightField('privacy');
    }
    
    // Проверяем email если указан
    const email = form.querySelector('#email').value.trim();
    if (email && !isValidEmail(email)) {
        errors.push('Укажите корректный email адрес');
        highlightField('email');
    }
    
    if (errors.length > 0) {
        showValidationErrors(errors);
        return false;
    }
    
    return true;
}

// Отправка формы на сервер
async function submitContactForm(data) {
    const response = await fetch('/api/contact', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data)
    });
    
    const result = await response.json();
    
    return {
        success: response.ok,
        message: result.message,
        data: result
    };
}

// Инициализация маски для телефона
function initPhoneMask() {
    const phoneInput = document.getElementById('phone');
    if (!phoneInput) return;
    
    phoneInput.addEventListener('input', function(e) {
        let value = e.target.value.replace(/\D/g, '');
        
        // Автоматически добавляем +7 если номер начинается с 8
        if (value.startsWith('8')) {
            value = '7' + value.slice(1);
        }
        
        // Форматируем номер
        if (value.startsWith('7')) {
            value = value.slice(1);
            if (value.length >= 1) {
                e.target.value = `+7 (${value.slice(0, 3)}${value.length > 3 ? ') ' + value.slice(3, 6) : ''}${value.length > 6 ? '-' + value.slice(6, 8) : ''}${value.length > 8 ? '-' + value.slice(8, 10) : ''}`;
            } else {
                e.target.value = '+7';
            }
        } else if (value) {
            e.target.value = '+7 (' + value.slice(0, 3) + (value.length > 3 ? ') ' + value.slice(3, 6) : '') + (value.length > 6 ? '-' + value.slice(6, 8) : '') + (value.length > 8 ? '-' + value.slice(8, 10) : '');
        }
    });
    
    // Устанавливаем начальное значение
    if (!phoneInput.value) {
        phoneInput.value = '+7';
    }
    
    // Не даем удалить +7
    phoneInput.addEventListener('keydown', function(e) {
        if (e.target.value.length <= 2 && (e.key === 'Backspace' || e.key === 'Delete')) {
            e.preventDefault();
        }
    });
}

// Инициализация live валидации
function initFormValidation() {
    const form = document.getElementById('contactForm');
    if (!form) return;
    
    // Валидация в реальном времени
    const inputs = form.querySelectorAll('input, textarea, select');
    inputs.forEach(input => {
        input.addEventListener('blur', function() {
            validateField(this);
        });
        
        input.addEventListener('input', function() {
            clearFieldError(this);
        });
    });
}

// Валидация отдельного поля
function validateField(field) {
    const value = field.value.trim();
    const fieldName = field.name;
    
    switch (fieldName) {
        case 'name':
            if (!value) {
                showFieldError(field, 'Укажите ваше имя');
            } else if (value.length < 2) {
                showFieldError(field, 'Имя должно содержать минимум 2 символа');
            }
            break;
            
        case 'phone':
            if (!value) {
                showFieldError(field, 'Укажите номер телефона');
            } else if (!isValidPhone(value)) {
                showFieldError(field, 'Укажите корректный номер телефона');
            }
            break;
            
        case 'email':
            if (value && !isValidEmail(value)) {
                showFieldError(field, 'Укажите корректный email адрес');
            }
            break;
    }
}

// Показ ошибки поля
function showFieldError(field, message) {
    clearFieldError(field);
    
    field.classList.add('error');
    
    const errorDiv = document.createElement('div');
    errorDiv.className = 'field-error';
    errorDiv.textContent = message;
    
    field.parentNode.appendChild(errorDiv);
}

// Очистка ошибки поля
function clearFieldError(field) {
    field.classList.remove('error');
    
    const errorDiv = field.parentNode.querySelector('.field-error');
    if (errorDiv) {
        errorDiv.remove();
    }
}

// Подсветка поля с ошибкой
function highlightField(fieldId) {
    const field = document.getElementById(fieldId);
    if (field) {
        field.focus();
        field.classList.add('error');
        
        setTimeout(() => {
            field.classList.remove('error');
        }, 3000);
    }
}

// Показ ошибок валидации (список точками)
function showValidationErrors(errors) {
    const html = 'Пожалуйста, исправьте ошибки:<br>• ' + errors.join('<br>• ');
    showToast(html, 'error', 'Ошибка');
}

// Показ сообщения об успешной отправке
function showSuccessMessage() {
    openSuccessModal();
}

// Единичная ошибка (например, при fetch)
function showErrorMessage(message) {
    showToast(message, 'error', 'Ошибка');
}

// Сброс формы
function resetContactForm() {
    const form = document.getElementById('contactForm');
    if (form) {
        form.reset();
        
        // Восстанавливаем маску телефона
        const phoneInput = document.getElementById('phone');
        if (phoneInput) {
            phoneInput.value = '+7';
        }
        
        // Очищаем все ошибки
        const errorFields = form.querySelectorAll('.error');
        errorFields.forEach(field => clearFieldError(field));
    }
}

// Установка состояния отправки
function setSubmitState(loading) {
    const submitBtn = document.querySelector('#contactForm button[type="submit"]');
    const btnText = submitBtn?.querySelector('.btn-text');
    const btnLoading = submitBtn?.querySelector('.btn-loading');
    
    if (loading) {
        if (submitBtn) submitBtn.disabled = true;
        if (btnText) btnText.classList.add('hidden');
        if (btnLoading) btnLoading.classList.remove('hidden');
    } else {
        if (submitBtn) submitBtn.disabled = false;
        if (btnText) btnText.classList.remove('hidden');
        if (btnLoading) btnLoading.classList.add('hidden');
    }
}

// Утилитарные функции валидации
function isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}

function isValidPhone(phone) {
    // Убираем все символы кроме цифр
    const cleanPhone = phone.replace(/\D/g, '');
    
    // Проверяем длину (должно быть 11 цифр для России)
    return cleanPhone.length === 11 && cleanPhone.startsWith('7');
}

// Тосты
function ensureToastWrap() {
    let wrap = document.querySelector('.toast-wrap');
    if (!wrap) {
        wrap = document.createElement('div');
        wrap.className = 'toast-wrap';
        document.body.appendChild(wrap);
    }
    return wrap;
}

function showToast(message, type = 'success', title = null, timeoutMs = 4200) {
    const wrap = ensureToastWrap();
    const el = document.createElement('div');
    el.className = `toast toast--${type}`;
    el.style.position = 'relative';

    el.innerHTML = `
        <button class="toast-close" aria-label="Закрыть">&times;</button>
        ${title ? `<div class="toast-title">${title}</div>` : ''}
        <div class="toast-text">${message}</div>
    `;

    const close = () => {
        if (!el.parentNode) return;
        el.parentNode.removeChild(el);
    };

    el.querySelector('.toast-close').addEventListener('click', close);
    wrap.appendChild(el);
    if (timeoutMs) setTimeout(close, timeoutMs);
}

// Открытие модалки успешной отправки
function openSuccessModal() {
    const modal = document.getElementById('contact-success-modal');
    if (!modal) return;

    const okBtn = document.getElementById('contact-success-ok');
    const closeBtn = modal.querySelector('.modal-close');

    // показать
    modal.classList.remove('hidden');
    modal.setAttribute('aria-hidden', 'false');

    // обработчики закрытия
    function closeModal() {
        modal.classList.add('hidden');
        modal.setAttribute('aria-hidden', 'true');
    }

    if (okBtn) okBtn.onclick = closeModal;
    if (closeBtn) closeBtn.onclick = closeModal;

    modal.addEventListener('click', (e) => {
        if (e.target === modal) closeModal();
    });

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && !modal.classList.contains('hidden')) closeModal();
    });
}
