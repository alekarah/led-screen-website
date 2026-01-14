// Admin Prices - управление ценами в админке

// ============== МОДАЛЬНЫЕ ОКНА ==============

function openCreatePriceModal() {
    const modal = document.getElementById('createPriceModal');
    if (!modal) return;

    // Очищаем форму
    const form = document.getElementById('createPriceForm');
    if (form) form.reset();

    // Сбрасываем характеристики
    const specsContainer = document.getElementById('createSpecifications');
    if (specsContainer) specsContainer.innerHTML = '';

    // Скрываем блок характеристик
    const specsBlock = document.getElementById('createSpecificationsBlock');
    if (specsBlock) specsBlock.style.display = 'none';

    modal.classList.add('active');
    document.body.style.overflow = 'hidden';
}

function openEditPriceModal(id) {
    const modal = document.getElementById('editPriceModal');
    if (!modal) return;

    // Загружаем данные позиции
    fetch(`/admin/prices/${id}`)
        .then(response => response.json())
        .then(data => {
            if (!data.price_item) {
                showAdminMessage('Позиция не найдена', 'error');
                return;
            }

            const priceItem = data.price_item;

            // Заполняем форму
            document.getElementById('editPriceId').value = priceItem.id;
            document.getElementById('editTitle').value = priceItem.title || '';
            document.getElementById('editDescription').value = priceItem.description || '';
            document.getElementById('editPriceFrom').value = priceItem.price_from || 0;
            document.getElementById('editHasSpecifications').checked = priceItem.has_specifications || false;
            document.getElementById('editIsActive').checked = priceItem.is_active !== false;

            // Отображаем текущее изображение с кнопками управления
            if (typeof displayPriceImage === 'function') {
                displayPriceImage(priceItem);
            }

            // Загружаем характеристики
            loadSpecifications('edit', priceItem.specifications || []);

            // Показываем/скрываем блок характеристик
            const specsBlock = document.getElementById('editSpecificationsBlock');
            if (specsBlock) {
                specsBlock.style.display = priceItem.has_specifications ? 'block' : 'none';
            }

            // Отображаем изображение с кнопками управления
            if (typeof displayPriceImage === 'function') {
                displayPriceImage(priceItem);
            }

            modal.classList.add('active');
            document.body.style.overflow = 'hidden';
        })
        .catch(error => {
            console.error('Ошибка загрузки позиции:', error);
            showAdminMessage('Ошибка загрузки данных', 'error');
        });
}

function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (!modal) return;

    modal.classList.remove('active');
    document.body.style.overflow = '';
}

// ============== РАБОТА С ХАРАКТЕРИСТИКАМИ ==============

function loadSpecifications(prefix, specifications) {
    const container = document.getElementById(`${prefix}Specifications`);
    if (!container) return;

    container.innerHTML = '';

    if (!specifications || specifications.length === 0) return;

    // Группируем характеристики по группам
    const groups = {};
    specifications.forEach(spec => {
        if (!groups[spec.spec_group]) {
            groups[spec.spec_group] = [];
        }
        groups[spec.spec_group].push(spec);
    });

    // Создаем блоки для каждой группы
    Object.keys(groups).forEach(groupName => {
        addSpecificationGroup(prefix, groupName, groups[groupName]);
    });
}

function addSpecificationGroup(prefix, groupName = '', specs = []) {
    const container = document.getElementById(`${prefix}Specifications`);
    if (!container) return;

    const groupDiv = document.createElement('div');
    groupDiv.className = 'spec-group';

    const groupHtml = `
        <div class="spec-group-header">
            <input type="text"
                   class="form-input spec-group-name"
                   placeholder="Название группы (напр. Параметры экрана)"
                   value="${groupName || ''}"
                   required>
            <button type="button" class="btn-icon btn-remove-group" onclick="removeSpecGroup(this)" title="Удалить группу">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M6 2L4 4H0V6H20V4H16L14 2H6ZM3 8V18C3 19.1 3.9 20 5 20H15C16.1 20 17 19.1 17 18V8H3Z"/>
                </svg>
            </button>
        </div>
        <div class="spec-rows"></div>
        <button type="button" class="btn btn-small" onclick="addSpecificationRow(this)">+ Добавить характеристику</button>
    `;

    groupDiv.innerHTML = groupHtml;
    container.appendChild(groupDiv);

    // Добавляем существующие характеристики или одну пустую
    const rowsContainer = groupDiv.querySelector('.spec-rows');
    if (specs && specs.length > 0) {
        specs.forEach(spec => {
            addSpecRow(rowsContainer, spec.spec_key || '', spec.spec_value || '');
        });
    } else {
        addSpecRow(rowsContainer, '', '');
    }
}

function addSpecificationRow(button) {
    const group = button.closest('.spec-group');
    const rowsContainer = group.querySelector('.spec-rows');
    addSpecRow(rowsContainer, '', '');
}

function addSpecRow(container, key = '', value = '') {
    const row = document.createElement('div');
    row.className = 'spec-row';
    row.innerHTML = `
        <input type="text" class="form-input spec-key" placeholder="Характеристика" value="${key || ''}" required>
        <input type="text" class="form-input spec-value" placeholder="Значение" value="${value || ''}" required>
        <button type="button" class="btn-icon btn-remove-row" onclick="removeSpecRow(this)" title="Удалить">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <path d="M4 6h8v6c0 1.1-.9 2-2 2H6c-1.1 0-2-.9-2-2V6zm3-4h2l1 1h3v2H3V3h3l1-1z"/>
            </svg>
        </button>
    `;
    container.appendChild(row);
}

function removeSpecGroup(button) {
    button.closest('.spec-group').remove();
}

function removeSpecRow(button) {
    button.closest('.spec-row').remove();
}

function toggleSpecifications(prefix) {
    const checkbox = document.getElementById(`${prefix}HasSpecifications`);
    const block = document.getElementById(`${prefix}SpecificationsBlock`);

    if (!checkbox || !block) return;

    block.style.display = checkbox.checked ? 'block' : 'none';
}

// ============== ПРЕВЬЮ ИЗОБРАЖЕНИЙ ==============

function previewImage(input, previewId) {
    const preview = document.getElementById(previewId);
    if (!preview || !input.files || !input.files[0]) return;

    const reader = new FileReader();
    reader.onload = function(e) {
        preview.innerHTML = `<img src="${e.target.result}" alt="Превью" style="max-width: 100%; max-height: 200px; border-radius: 8px;">`;
        preview.style.display = 'block';
    };
    reader.readAsDataURL(input.files[0]);
}

// ============== СОЗДАНИЕ ПОЗИЦИИ ==============

function submitCreatePrice(event) {
    event.preventDefault();

    const form = event.target;
    const formData = new FormData(form);

    // Собираем характеристики
    const hasSpecs = document.getElementById('createHasSpecifications').checked;
    if (hasSpecs) {
        const specs = collectSpecifications('create');
        formData.append('specifications', JSON.stringify(specs));
    }

    // Отправляем форму
    fetch('/admin/prices', {
        method: 'POST',
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        if (data.success && data.price_id) {
            showAdminMessage('Позиция создана. Открываю редактирование…', 'success');
            const priceId = data.price_id;
            closeModal('createPriceModal');
            // Открываем модалку редактирования для добавления изображения
            setTimeout(() => {
                openEditPriceModal(priceId);
            }, 100);
        } else {
            showAdminMessage(data.error || 'Ошибка создания позиции', 'error');
        }
    })
    .catch(error => {
        console.error('Ошибка:', error);
        showAdminMessage('Ошибка создания позиции', 'error');
    });
}

// ============== ОБНОВЛЕНИЕ ПОЗИЦИИ ==============

function submitEditPrice(event) {
    event.preventDefault();

    const form = event.target;
    const id = document.getElementById('editPriceId').value;
    const formData = new FormData(form);

    // Собираем характеристики
    const hasSpecs = document.getElementById('editHasSpecifications').checked;
    if (hasSpecs) {
        const specs = collectSpecifications('edit');
        formData.append('specifications', JSON.stringify(specs));
    }

    // Дебаг: логируем что отправляется
    console.log('=== Отправка формы редактирования ===');
    for (let [key, value] of formData.entries()) {
        console.log(`${key}:`, value);
    }

    // Отправляем форму
    fetch(`/admin/prices/${id}/update`, {
        method: 'POST',
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showAdminMessage(data.message || 'Позиция успешно обновлена', 'success');
            closeModal('editPriceModal');
            setTimeout(() => location.reload(), 500);
        } else {
            showAdminMessage(data.error || 'Ошибка обновления позиции', 'error');
        }
    })
    .catch(error => {
        console.error('Ошибка:', error);
        showAdminMessage('Ошибка обновления позиции', 'error');
    });
}

// ============== СБОР ХАРАКТЕРИСТИК ==============

function collectSpecifications(prefix) {
    const container = document.getElementById(`${prefix}Specifications`);
    if (!container) return [];

    const specs = [];
    const groups = container.querySelectorAll('.spec-group');

    groups.forEach((group, groupIndex) => {
        const groupName = group.querySelector('.spec-group-name').value.trim();
        if (!groupName) return;

        const rows = group.querySelectorAll('.spec-row');
        rows.forEach((row, rowIndex) => {
            const key = row.querySelector('.spec-key').value.trim();
            const value = row.querySelector('.spec-value').value.trim();

            if (key && value) {
                specs.push({
                    group: groupName,
                    key: key,
                    value: value,
                    order: groupIndex * 1000 + rowIndex
                });
            }
        });
    });

    return specs;
}

// ============== УДАЛЕНИЕ ПОЗИЦИИ ==============

function deletePrice(id, title) {
    if (!confirm(`Вы уверены, что хотите удалить позицию "${title}"?`)) return;

    fetch(`/admin/prices/${id}`, {
        method: 'DELETE'
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showAdminMessage(data.message || 'Позиция удалена', 'success');
            setTimeout(() => location.reload(), 500);
        } else {
            showAdminMessage(data.error || 'Ошибка удаления', 'error');
        }
    })
    .catch(error => {
        console.error('Ошибка:', error);
        showAdminMessage('Ошибка удаления позиции', 'error');
    });
}

// ============== DRAG & DROP СОРТИРОВКА ==============

function initPriceSorting() {
    const container = document.getElementById('sortable-prices');
    if (!container || typeof Sortable === 'undefined') return;

    new Sortable(container, {
        animation: 150,
        handle: '.drag-handle',
        ghostClass: 'sortable-ghost',
        onEnd: function() {
            savePriceOrder();
        }
    });
}

function savePriceOrder() {
    const container = document.getElementById('sortable-prices');
    if (!container) return;

    const items = container.querySelectorAll('.price-item');
    const ids = Array.from(items).map(item => parseInt(item.dataset.priceId));

    fetch('/admin/prices/sort', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ ids: ids })
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showAdminMessage('Порядок обновлен', 'success');
        } else {
            showAdminMessage('Ошибка обновления порядка', 'error');
        }
    })
    .catch(error => {
        console.error('Ошибка:', error);
        showAdminMessage('Ошибка обновления порядка', 'error');
    });
}

// ============== ИНИЦИАЛИЗАЦИЯ ==============

document.addEventListener('DOMContentLoaded', function() {
    // Инициализация drag & drop сортировки
    initPriceSorting();

    // Обработчики для открытия модалок
    const createBtn = document.getElementById('openCreatePriceModal');
    if (createBtn) createBtn.onclick = openCreatePriceModal;

    // Обработчики для закрытия модалок
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', function(e) {
            if (e.target === modal) {
                closeModal(modal.id);
            }
        });
    });

    // ESC для закрытия модалок
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape') {
            document.querySelectorAll('.modal.active').forEach(modal => {
                closeModal(modal.id);
            });
        }
    });

    // Обработчики для чекбоксов характеристик
    const createHasSpecs = document.getElementById('createHasSpecifications');
    if (createHasSpecs) {
        createHasSpecs.addEventListener('change', () => toggleSpecifications('create'));
    }

    const editHasSpecs = document.getElementById('editHasSpecifications');
    if (editHasSpecs) {
        editHasSpecs.addEventListener('change', () => toggleSpecifications('edit'));
    }

    // Обработчики для форм
    const createForm = document.getElementById('createPriceForm');
    if (createForm) createForm.onsubmit = submitCreatePrice;

    const editForm = document.getElementById('editPriceForm');
    if (editForm) editForm.onsubmit = submitEditPrice;

    // Кнопки добавления групп характеристик
    const addCreateGroupBtn = document.getElementById('addCreateSpecGroup');
    if (addCreateGroupBtn) {
        addCreateGroupBtn.onclick = () => addSpecificationGroup('create');
    }

    const addEditGroupBtn = document.getElementById('addEditSpecGroup');
    if (addEditGroupBtn) {
        addEditGroupBtn.onclick = () => addSpecificationGroup('edit');
    }
});
