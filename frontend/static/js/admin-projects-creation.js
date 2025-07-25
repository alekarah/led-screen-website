// Создание новых проектов в админке

// Создание нового проекта
function initProjectCreation() {
    const createProjectForm = document.getElementById('createProjectForm');
    if (!createProjectForm) return;

    createProjectForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        if (!validateProjectForm(createProjectForm)) {
            showAdminMessage('Заполните все обязательные поля', 'error');
            return;
        }

        setSubmitButtonState(createProjectForm, true);

        try {
            const formData = new FormData(createProjectForm);

            const response = await fetch('/admin/projects', {
                method: 'POST',
                body: formData
            });

            const result = await response.json();

            if (response.ok && result.project_id) {
                showUploadSection(result.project_id);
                resetForm('createProjectForm');
                showAdminMessage('Проект успешно создан! Теперь можете загрузить изображения.', 'success');
                // location.reload(); // Не перезагружаем!
            } else {
                throw new Error(result.error || 'Ошибка создания проекта');
            }
        } catch (error) {
            console.error('Ошибка создания проекта:', error);
            showAdminMessage('Ошибка при создании проекта: ' + error.message, 'error');
        } finally {
            setSubmitButtonState(createProjectForm, false);
        }
    });
}

// Валидация формы создания проекта
function validateProjectForm(form) {
    const requiredFields = ['title'];

    for (const fieldName of requiredFields) {
        const field = form.querySelector(`[name="${fieldName}"]`);
        if (!field || !field.value.trim()) {
            highlightInvalidField(field);
            return false;
        }
    }

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

    field.addEventListener('input', function () {
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
        uploadSection.scrollIntoView({
            behavior: 'smooth',
            block: 'center'
        });
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

        if (!validateImageFiles(fileInput.files)) {
            return;
        }

        showUploadProgress(true);

        try {
            const formData = new FormData(uploadForm);

            const response = await fetch('/admin/upload-images', {
                method: 'POST',
                body: formData
            });

            const result = await response.json();

            if (response.ok) {
                hideUploadSection();
                showAdminMessage('Изображения успешно загружены!', 'success');
                resetForm('uploadForm');
                closeCreateProjectModal();
                // setTimeout(() => location.reload(), 300); // Лучше не перезагружать, чтобы пользователь мог дальше работать
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
        if (file.size > maxFileSize) {
            showAdminMessage(`Файл "${file.name}" слишком большой. Максимальный размер: 10MB`, 'error');
            return false;
        }
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

function openCreateProjectModal() {
    document.getElementById('createProjectModal').style.display = 'block';
    document.body.style.overflow = 'hidden';
}
function closeCreateProjectModal() {
    document.getElementById('createProjectModal').style.display = 'none';
    document.body.style.overflow = 'auto';
    resetForm('createProjectForm');
    resetForm('uploadForm');
    hideUploadSection();
}

document.addEventListener('DOMContentLoaded', function () {
    initProjectCreation();
    initImageUpload();
    const btn = document.getElementById('openCreateProjectModal');
    if (btn) {
        btn.onclick = openCreateProjectModal;
    }
});

window.initProjectCreation = initProjectCreation;
window.initImageUpload = initImageUpload;
window.showUploadSection = showUploadSection;