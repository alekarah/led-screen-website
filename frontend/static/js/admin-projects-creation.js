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

// Подсчет количества выбранных проектов на главной
function countFeaturedProjects() {
    // Считаем все проекты с галочкой featured в базе через DOM
    const featuredItems = document.querySelectorAll('.project-item .featured-yes');
    return featuredItems.length;
}

// Обновление предупреждения и счетчика для модального окна создания
function updateFeaturedWarning(checkboxId, modalType) {
    const checkbox = document.getElementById(checkboxId);
    if (!checkbox) return;

    const modal = checkbox.closest('.modal');
    if (!modal) return;

    let warning = modal.querySelector('.featured-warning');
    const label = modal.querySelector(`label[for="${checkboxId}"]`);

    // Подсчитываем текущее количество featured проектов
    let currentCount = countFeaturedProjects();

    // Если в модале редактирования, и чекбокс был уже выбран при открытии,
    // не учитываем текущий проект в счетчике (он уже учтен)
    if (modalType === 'edit') {
        const projectId = document.getElementById('edit_project_id');
        if (projectId && projectId.value) {
            // Проверяем начальное состояние чекбокса
            const initiallyChecked = checkbox.dataset.initiallyChecked === 'true';
            if (initiallyChecked && !checkbox.checked) {
                currentCount--; // Проект был featured, но мы сняли галочку
            } else if (!initiallyChecked && checkbox.checked) {
                currentCount++; // Проект не был featured, но мы поставили галочку
            }
        }
    } else {
        // Для создания - просто прибавляем 1 если галочка стоит
        if (checkbox.checked) {
            currentCount++;
        }
    }

    // Обновляем текст в label
    if (label) {
        const baseText = '⭐ Показать на главной странице';
        label.textContent = `${baseText} (выбрано: ${currentCount} из 6)`;
    }

    // Показываем/скрываем предупреждение
    if (currentCount > 6) {
        if (!warning) {
            // Создаем предупреждение если его нет
            warning = document.createElement('div');
            warning.className = 'featured-warning';
            warning.innerHTML = '⚠️ Будут показаны первые 6 проектов по порядку сортировки';

            // Вставляем после чекбокса
            const featuredCheckbox = modal.querySelector('.featured-checkbox');
            if (featuredCheckbox) {
                featuredCheckbox.insertAdjacentElement('afterend', warning);
            }
        }
        warning.style.display = 'block';
    } else {
        if (warning) {
            warning.style.display = 'none';
        }
    }
}

// Инициализация обработчиков для чекбокса featured в модальном окне создания
function initFeaturedCheckboxHandlers() {
    // Для модального окна создания
    const createCheckbox = document.getElementById('featured');
    if (createCheckbox) {
        createCheckbox.addEventListener('change', function() {
            updateFeaturedWarning('featured', 'create');
        });

        // Обновляем при открытии модального окна
        const observer = new MutationObserver(function(mutations) {
            mutations.forEach(function(mutation) {
                const modal = document.getElementById('createProjectModal');
                if (modal && modal.style.display === 'block') {
                    updateFeaturedWarning('featured', 'create');
                }
            });
        });

        const createModal = document.getElementById('createProjectModal');
        if (createModal) {
            observer.observe(createModal, { attributes: true, attributeFilter: ['style'] });
        }
    }
}

// Инициализируем обработчики при загрузке
document.addEventListener('DOMContentLoaded', initFeaturedCheckboxHandlers);

window.initProjectCreation = initProjectCreation;
window.updateFeaturedWarning = updateFeaturedWarning;