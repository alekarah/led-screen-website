// Создание новых проектов в админке

// Создание нового проекта
function initProjectCreation() {
    const projectForm = document.getElementById('projectForm');
    if (!projectForm) return;

    projectForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        // Валидация формы
        if (!validateProjectForm(projectForm)) {
            showAdminMessage('Заполните все обязательные поля', 'error');
            return;
        }
        
        // Показываем индикатор загрузки
        setSubmitButtonState(projectForm, true);
        
        try {
            const formData = new FormData(projectForm);
            
            const response = await fetch('/admin/projects', {
                method: 'POST',
                body: formData
            });
            
            const result = await response.json();
            
            if (response.ok && result.project_id) {
                // Показываем форму загрузки изображений
                showUploadSection(result.project_id);
                
                // Очищаем форму
                resetForm('projectForm');
                
                showAdminMessage('Проект успешно создан! Теперь можете загрузить изображения.', 'success');
                location.reload();
            } else {
                throw new Error(result.error || 'Ошибка создания проекта');
            }
        } catch (error) {
            console.error('Ошибка создания проекта:', error);
            showAdminMessage('Ошибка при создании проекта: ' + error.message, 'error');
        } finally {
            setSubmitButtonState(projectForm, false);
        }
    });
}

// Валидация формы создания проекта
function validateProjectForm(form) {
    const requiredFields = ['title'];
    
    for (const fieldName of requiredFields) {
        const field = form.querySelector(`[name="${fieldName}"]`);
        if (!field || !field.value.trim()) {
            // Подсвечиваем невалидное поле
            highlightInvalidField(field);
            return false;
        }
    }
    
    // Проверяем специфические правила
    const title = form.querySelector('[name="title"]').value.trim();
    if (title.length < 3) {
        showAdminMessage('Название проекта должно содержать минимум 3 символа', 'error');
        return false;
    }
    
    return true;
}

// Подсветка невалидного поля
function highlightInvalidField(field) {
    if (!field) return;
    
    field.style.borderColor = '#dc3545';
    field.focus();
    
    // Убираем подсветку при вводе
    field.addEventListener('input', function() {
        field.style.borderColor = '';
    }, { once: true });
}

// Показ секции загрузки изображений
function showUploadSection(projectId) {
    const uploadSection = document.getElementById('uploadSection');
    const projectIdInput = document.getElementById('project_id');
    
    if (uploadSection && projectIdInput) {
        projectIdInput.value = projectId;
        uploadSection.classList.remove('hidden');
        
        // Прокручиваем к секции загрузки с анимацией
        uploadSection.scrollIntoView({ 
            behavior: 'smooth',
            block: 'center'
        });
        
        // Добавляем визуальный эффект
        uploadSection.style.animation = 'fadeIn 0.5s ease-in';
    }
}

// Загрузка изображений для нового проекта
function initImageUpload() {
    const uploadForm = document.getElementById('uploadForm');
    if (!uploadForm) return;

    uploadForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const fileInput = document.getElementById('images');
        if (!fileInput.files.length) {
            showAdminMessage('Выберите изображения для загрузки', 'error');
            return;
        }
        
        // Валидация файлов
        if (!validateImageFiles(fileInput.files)) {
            return;
        }
        
        // Показываем прогресс загрузки
        showUploadProgress(true);
        
        try {
            const formData = new FormData(uploadForm);
            
            const response = await fetch('/admin/upload-images', {
                method: 'POST',
                body: formData
            });
            
            const result = await response.json();
            
            if (response.ok) {
                // Скрываем форму загрузки
                hideUploadSection();
                
                showAdminMessage('Изображения успешно загружены! Страница будет обновлена.', 'success');
                
                // Перезагружаем страницу через 2 секунды
                setTimeout(() => location.reload(), 300);
            } else {
                throw new Error(result.error || 'Ошибка загрузки изображений');
            }
        } catch (error) {
            console.error('Ошибка загрузки изображений:', error);
            showAdminMessage('Ошибка при загрузке изображений: ' + error.message, 'error');
        } finally {
            showUploadProgress(false);
        }
    });
}

// Валидация загружаемых изображений
function validateImageFiles(files) {
    const maxFileSize = 10 * 1024 * 1024; // 10MB
    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp'];
    
    for (const file of files) {
        // Проверяем размер файла
        if (file.size > maxFileSize) {
            showAdminMessage(`Файл "${file.name}" слишком большой. Максимальный размер: 10MB`, 'error');
            return false;
        }
        
        // Проверяем тип файла
        if (!allowedTypes.includes(file.type)) {
            showAdminMessage(`Файл "${file.name}" имеет неподдерживаемый формат. Разрешены: JPEG, PNG, WebP`, 'error');
            return false;
        }
    }
    
    return true;
}

// Скрытие секции загрузки
function hideUploadSection() {
    const uploadSection = document.getElementById('uploadSection');
    if (uploadSection) {
        uploadSection.style.animation = 'fadeOut 0.3s ease-out';
        setTimeout(() => {
            uploadSection.classList.add('hidden');
            uploadSection.style.animation = '';
        }, 300);
    }
}

// Показ/скрытие прогресса загрузки
function showUploadProgress(show) {
    const submitBtn = document.querySelector('#uploadForm button[type="submit"]');
    const fileInput = document.getElementById('images');
    
    if (show) {
        if (submitBtn) {
            submitBtn.disabled = true;
            submitBtn.textContent = 'Загружаем...';
        }
        if (fileInput) fileInput.disabled = true;
    } else {
        if (submitBtn) {
            submitBtn.disabled = false;
            submitBtn.textContent = 'Загрузить изображения';
        }
        if (fileInput) fileInput.disabled = false;
    }
}

// Установка состояния кнопки отправки
function setSubmitButtonState(form, loading) {
    const submitBtn = form.querySelector('button[type="submit"]');
    if (!submitBtn) return;
    
    if (loading) {
        submitBtn.disabled = true;
        submitBtn.textContent = 'Создаем...';
        submitBtn.style.opacity = '0.7';
    } else {
        submitBtn.disabled = false;
        submitBtn.textContent = 'Создать проект';
        submitBtn.style.opacity = '1';
    }
}

// Экспорт функций
window.initProjectCreation = initProjectCreation;
window.initImageUpload = initImageUpload;
window.showUploadSection = showUploadSection;

// ПРИНУДИТЕЛЬНЫЙ ВЫЗОВ ФУНКЦИИ
document.addEventListener('DOMContentLoaded', function() {
    console.log('🔧 Принудительная инициализация создания проектов...');
    initProjectCreation();
    console.log('✅ initProjectCreation вызван принудительно');
});