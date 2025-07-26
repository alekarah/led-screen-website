// Drag & Drop сортировка проектов

let sortableInstance = null;

// Инициализация сортировки
function initProjectSorting() {
    const sortableContainer = document.getElementById('sortable-projects');
    if (!sortableContainer) return;

    sortableInstance = Sortable.create(sortableContainer, {
        handle: '.drag-handle', // Только за иконку можно перетаскивать
        animation: 150, // Плавная анимация
        ghostClass: 'sortable-ghost', // Класс для призрака
        chosenClass: 'sortable-chosen', // Класс для выбранного элемента
        dragClass: 'sortable-drag', // Класс во время перетаскивания
        
        // Функция срабатывает при изменении порядка
        onEnd: function(evt) {
            const projectId = evt.item.getAttribute('data-project-id');
            const newIndex = evt.newIndex;
            const oldIndex = evt.oldIndex;
                       
            // Сохраняем новый порядок на сервере
            updateProjectOrder(projectId, newIndex);
        },
        
        // Визуальная обратная связь
        onStart: function(evt) {
            showMessage('Перетащите проект в нужную позицию', 'success');
        }
    });
}

// Сохранение нового порядка на сервере
async function updateProjectOrder(projectId, newPosition) {
    try {
        const formData = new FormData();
        formData.append('position', newPosition);
        
        const response = await fetch(`/admin/projects/${projectId}/reorder`, {
            method: 'POST',
            body: formData
        });
        
        const result = await response.json();
        
        if (response.ok) {
            showMessage('Порядок проектов обновлен', 'success');
            
            // Обновляем порядок всех проектов
            await updateAllProjectsOrder();
        } else {
            showMessage(result.error || 'Ошибка изменения порядка', 'error');
            
            // Возвращаем элемент на место при ошибке
            location.reload();
        }
    } catch (error) {
        console.error('Ошибка обновления порядка:', error);
        showMessage('Ошибка сети: ' + error.message, 'error');
        
        // Возвращаем элемент на место при ошибке
        location.reload();
    }
}

// Обновление порядка всех проектов
async function updateAllProjectsOrder() {
    const projectItems = document.querySelectorAll('#sortable-projects .project-item');
    const orderData = [];
    
    projectItems.forEach((item, index) => {
        const projectId = item.getAttribute('data-project-id');
        orderData.push({
            id: parseInt(projectId),
            sort_order: index
        });
    });
    
    try {
        const response = await fetch('/admin/projects/bulk-reorder', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ projects: orderData })
        });
        
        if (!response.ok) {
            console.error('Ошибка массового обновления порядка');
        }
    } catch (error) {
        console.error('Ошибка массового обновления:', error);
    }
}

// Сброс порядка к алфавитному
async function resetProjectOrder() {
    if (!confirm('Сбросить порядок проектов к алфавитному?')) return;
    
    try {
        const response = await fetch('/admin/projects/reset-order', {
            method: 'POST'
        });
        
        const result = await response.json();
        
        if (response.ok) {
            showMessage('Порядок проектов сброшен', 'success');
            setTimeout(() => location.reload(), 1000);
        } else {
            showMessage(result.error || 'Ошибка сброса порядка', 'error');
        }
    } catch (error) {
        showMessage('Ошибка: ' + error.message, 'error');
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    initProjectSorting();
});

// Глобальные функции
window.resetProjectOrder = resetProjectOrder;