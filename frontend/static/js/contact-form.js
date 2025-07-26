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

// Показ ошибок валидации
function showValidationErrors(errors) {
    const message = 'Пожалуйста, исправьте ошибки:\n• ' + errors.join('\n• ');
    showMessage(message, 'error');
}

// Показ сообщения об успешной отправке
function showSuccessMessage() {
    const message = `
        ✅ Заявка успешно отправлена!
        
        Мы свяжемся с вами в течение 30 минут.
        
        Также вы можете связаться с нами напрямую:
        📞 +7 (921) 429-17-02
    `;
    
    showMessage(message, 'success');
}

// Показ сообщения об ошибке
function showErrorMessage(message) {
    showMessage(message, 'error');
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

// Показ уведомлений
function showMessage(message, type) {   
    if (type === 'error') {
        alert('❌ ' + message);
    } else {
        alert('✅ ' + message);
    }
}