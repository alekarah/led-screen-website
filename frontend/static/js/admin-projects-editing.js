// Редактирование проектов в админке

// ПРИНУДИТЕЛЬНОЕ СОЗДАНИЕ ГЛОБАЛЬНЫХ ФУНКЦИЙ
console.log('🔧 Загрузка функций редактирования проектов...');

// Глобальная функция редактирования
window.editProject = async function(id) {
    console.log('Редактирование проекта:', id);
    
    if (!id) {
        showAdminMessage('Некорректный ID проекта', 'error');
        return;
    }
    
    try {
        // Загружаем данные проекта
        const response = await fetch(`/admin/projects/${id}`);
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        
        const data = await response.json();
        console.log('Данные проекта:', data);
        
        // Заполняем форму редактирования
        fillEditForm(data);
        
        // Показываем модальное окно
        const modal = document.getElementById('editModal');
        if (modal) {
            modal.style.display = 'block';
            document.body.style.overflow = 'hidden';
        }
        
    } catch (error) {
        console.error('Ошибка загрузки проекта:', error);
        showAdminMessage('Ошибка загрузки данных проекта: ' + error.message, 'error');
    }
};

// Глобальная функция удаления
window.deleteProject = async function(id) {
    console.log('Удаление проекта:', id);
    
    if (!id) {
        showAdminMessage('Некорректный ID проекта', 'error');
        return;
    }
    
    if (!confirm('Вы уверены, что хотите удалить этот проект? Все изображения также будут удалены.')) {
        return;
    }
    
    try {
        const response = await fetch(`/admin/projects/${id}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            const result = await response.json();
            throw new Error(result.error || `HTTP ${response.status}`);
        }
        
        const result = await response.json();
        
        showAdminMessage('Проект успешно удален', 'success');
        setTimeout(() => location.reload(), 300);
        
    } catch (error) {
        console.error('Ошибка удаления проекта:', error);
        showAdminMessage('Ошибка при удалении проекта: ' + error.message, 'error');
    }
};

// Глобальная функция закрытия модального окна
window.closeEditModal = function() {
    console.log('Закрытие модального окна редактирования');
    
    const modal = document.getElementById('editModal');
    if (modal) {
        modal.style.display = 'none';
        document.body.style.overflow = 'auto';
    }
    
    // Очищаем форму
    const form = document.getElementById('editProjectForm');
    if (form) {
        form.reset();
    }
    
    // Очищаем изображения
    const imagesDiv = document.getElementById('project_images');
    if (imagesDiv) {
        imagesDiv.innerHTML = '';
    }
    
    // Очищаем категории
    const categoriesDiv = document.getElementById('edit_categories');
    if (categoriesDiv) {
        categoriesDiv.innerHTML = '';
    }
};

// Заполнение формы редактирования
function fillEditForm(data) {
    const { project, categories } = data;
    
    if (!project) {
        showAdminMessage('Данные проекта не найдены', 'error');
        return;
    }
    
    // Основные поля
    const fields = {
        'edit_project_id': project.id,
        'edit_title': project.title || '',
        'edit_description': project.description || '',
        'edit_location': project.location || '',
        'edit_size': project.size || '',
        'edit_pixel_pitch': project.pixel_pitch || '',
        'edit_upload_project_id': project.id
    };
    
    // Заполняем текстовые поля
    Object.entries(fields).forEach(([fieldId, value]) => {
        const field = document.getElementById(fieldId);
        if (field) {
            field.value = value;
        }
    });
    
    // Заполняем чекбокс "Показать на главной"
    const featuredCheckbox = document.getElementById('edit_featured');
    if (featuredCheckbox) {
        featuredCheckbox.checked = project.featured || false;
    }
    
    // Заполняем категории
    fillEditCategories(categories, project.categories || []);
    
    // Заполняем изображения
    fillProjectImages(project.images || []);
    
    // ПЕРЕСОЗДАЕМ ФОРМЫ ЧТОБЫ УБРАТЬ СТАРЫЕ ОБРАБОТЧИКИ
    initEditForms();
}

// Заполнение категорий в форме редактирования
function fillEditCategories(allCategories, projectCategories) {
    const categoriesDiv = document.getElementById('edit_categories');
    if (!categoriesDiv) {
        console.warn('Контейнер категорий не найден');
        return;
    }
    
    categoriesDiv.innerHTML = '';
    
    if (!allCategories || !Array.isArray(allCategories)) {
        categoriesDiv.innerHTML = '<p>Категории не загружены</p>';
        return;
    }
    
    allCategories.forEach(category => {
        const isChecked = projectCategories.some(pc => pc.id === category.id);
        
        const checkboxHTML = `
            <div class="checkbox-item">
                <input type="checkbox" 
                       id="edit_cat_${category.id}" 
                       name="categories" 
                       value="${category.id}" 
                       ${isChecked ? 'checked' : ''}>
                <label for="edit_cat_${category.id}">${category.name}</label>
            </div>
        `;
        
        categoriesDiv.innerHTML += checkboxHTML;
    });
}

// Инициализация форм в модальном окне
function initEditForms() {
    // Пересоздаем форму редактирования
    const editForm = document.getElementById('editProjectForm');
    if (editForm) {
        const newEditForm = editForm.cloneNode(true);
        editForm.parentNode.replaceChild(newEditForm, editForm);
        
        newEditForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            console.log('📝 Обновление проекта...');
            
            const projectId = document.getElementById('edit_project_id').value;
            const title = this.querySelector('[name="title"]').value.trim();
            
            if (!projectId || !title) {
                showAdminMessage('Заполните обязательные поля', 'error');
                return;
            }
            
            try {
                const formData = new FormData(this);
                const response = await fetch(`/admin/projects/${projectId}/update`, {
                    method: 'POST',
                    body: formData
                });
                
                if (response.ok) {
                    showAdminMessage('Проект успешно обновлен', 'success');
                    closeEditModal();
                    setTimeout(() => location.reload(), 1000);
                } else {
                    const result = await response.json();
                    throw new Error(result.error || 'Ошибка сервера');
                }
            } catch (error) {
                console.error('Ошибка обновления:', error);
                showAdminMessage('Ошибка при обновлении проекта: ' + error.message, 'error');
            }
        });
        
        console.log('✅ Форма редактирования пересоздана');
    }
    
    // Пересоздаем форму загрузки изображений
    const uploadForm = document.getElementById('editUploadForm');
    if (uploadForm) {
        const newUploadForm = uploadForm.cloneNode(true);
        uploadForm.parentNode.replaceChild(newUploadForm, uploadForm);
        
        newUploadForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            console.log('📤 Загрузка изображений...');
            
            const fileInput = this.querySelector('input[type="file"]');
            if (!fileInput.files.length) {
                showAdminMessage('Выберите файлы для загрузки', 'error');
                return;
            }
            
            try {
                const formData = new FormData(this);
                const response = await fetch('/admin/upload-images', {
                    method: 'POST',
                    body: formData
                });
                
                if (response.ok) {
                    showAdminMessage('Изображения успешно загружены', 'success');
                    this.reset();
                    
                    // Обновляем изображения в модальном окне
                    const projectId = document.getElementById('edit_project_id').value;
                    if (projectId) {
                        setTimeout(async () => {
                            try {
                                const data = await fetch(`/admin/projects/${projectId}`).then(r => r.json());
                                fillProjectImages(data.project.images || []);
                            } catch (error) {
                                console.error('Ошибка обновления изображений:', error);
                            }
                        }, 500);
                    }
                } else {
                    const result = await response.json();
                    throw new Error(result.error || 'Ошибка загрузки');
                }
            } catch (error) {
                console.error('Ошибка загрузки:', error);
                showAdminMessage('Ошибка при загрузке изображений: ' + error.message, 'error');
            }
        });
        
        console.log('✅ Форма загрузки пересоздана');
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    console.log('✅ Функции редактирования проектов загружены');
    
    // Проверяем доступность функций
    console.log('Доступные функции:', {
        editProject: typeof window.editProject,
        deleteProject: typeof window.deleteProject,
        closeEditModal: typeof window.closeEditModal
    });
});

console.log('📋 Модуль редактирования проектов инициализирован');