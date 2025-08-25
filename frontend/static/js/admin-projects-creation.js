// Создание новых проектов в админке

// Создание нового проекта
function initProjectCreation() {
    const createProjectForm = document.getElementById('createProjectForm');
    if (!createProjectForm) return;
    if (createProjectForm.dataset.bound === '1') return;
    createProjectForm.dataset.bound = '1';

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
                showAdminMessage('Проект создан. Открываю редактирование…', 'success');
                const projectId = result.project_id;
                closeCreateProjectModal();
                setTimeout(() => {
                    if (typeof editProject === 'function') editProject(projectId);
                }, 100);
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
}

window.initProjectCreation = initProjectCreation;