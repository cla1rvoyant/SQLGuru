function openCreateModal(table) {
    currentMode = 'create';
    currentRecordId = '';
    document.getElementById('modalTitle').textContent = 'Создать запись';
    document.getElementById('submitButton').textContent = 'Создать';
    document.getElementById('formAction').value = 'create';
    document.getElementById('formId').value = '';
    document.getElementById('createForm').reset();
    updateFormFields(table);
    document.getElementById('createModal').style.display = 'block';
}

async function openEditModal(table, id) {
    currentMode = 'edit';
    currentRecordId = id;
    const response = await fetch(`/admin/get?table=${table}&id=${id}`);
    if (!response.ok) { alert('Ошибка загрузки данных'); return; }
    const data = await response.json();
    document.getElementById('modalTitle').textContent = 'Редактировать запись';
    document.getElementById('submitButton').textContent = 'Обновить';
    document.getElementById('formAction').value = 'update';
    document.getElementById('formId').value = id;
    document.getElementById('formTable').value = table;
    fillFormFields(table, data);
    document.getElementById('createModal').style.display = 'block';
}

function fillFormFields(table, data) {
    let html = '';
    switch(table) {
        case 'admins':
            html = `
                <div class="form-group"><label>Логин:</label>
                <input type="text" name="login" value="${data.login || ''}" required></div>
                <div class="form-group"><label>Пароль (пустое — без изменений):</label>
                <input type="password" name="password"></div>`;
            break;
        case 'topics':
            html = `
                <div class="form-group"><label>Название темы:</label>
                <input type="text" name="name" value="${data.name || ''}" required></div>`;
            break;
        case 'questions':
            html = `
                <div class="form-group"><label>Тема:</label>
                <select name="topic_id" required><option value="">Выберите тему</option></select></div>
                <div class="form-group"><label>Текст вопроса:</label>
                <textarea name="question_text" required>${data.question_text || ''}</textarea></div>
                <div class="form-group"><label>Правильный ответ:</label>
                <input type="text" name="correct_answer" value="${data.correct_answer || ''}" required></div>
                <div class="form-group"><label>Неправильный ответ 1:</label>
                <input type="text" name="wrong_answer1" value="${data.wrong_answer1 || ''}" required></div>
                <div class="form-group"><label>Неправильный ответ 2 (необязательно):</label>
                <input type="text" name="wrong_answer2" value="${data.wrong_answer2 || ''}"></div>
                <div class="form-group"><label>Неправильный ответ 3 (необязательно):</label>
                <input type="text" name="wrong_answer3" value="${data.wrong_answer3 || ''}"></div>`;
            break;
        case 'articles':
            html = `
                <div class="form-group"><label>Тема (topic_name):</label>
                <input type="text" name="topic_name" value="${data.topic_name || ''}" required></div>
                <div class="form-group"><label>Заголовок:</label>
                <input type="text" name="title" value="${data.title || ''}" required></div>
                <div class="form-group"><label>Содержимое (HTML):</label>
                <textarea name="content" rows="10" required>${data.content || ''}</textarea></div>`;
            break;
    }
    document.getElementById('formFields').innerHTML = html;
    if (table === 'questions') fillTopicsSelect(data.topic_id);
}

async function fillTopicsSelect(selectedTopicId = '') {
    const select = document.querySelector('select[name="topic_id"]');
    if (!select) return;
    const response = await fetch('/admin/topics');
    if (!response.ok) return;
    const topics = await response.json();
    select.innerHTML = '<option value="">Выберите тему</option>';
    topics.forEach(topic => {
        const option = document.createElement('option');
        option.value = topic.id;
        option.textContent = topic.name;
        if (topic.id == selectedTopicId) option.selected = true;
        select.appendChild(option);
    });
}

function updateFormFields(table) {
    let html = '';
    switch(table) {
        case 'admins':
            html = `
                <div class="form-group"><label>Логин:</label>
                <input type="text" name="login" required></div>
                <div class="form-group"><label>Пароль:</label>
                <input type="password" name="password" required></div>`;
            break;
        case 'topics':
            html = `
                <div class="form-group"><label>Название темы:</label>
                <input type="text" name="name" required></div>`;
            break;
        case 'questions':
            html = `
                <div class="form-group"><label>Тема:</label>
                <select name="topic_id" required><option value="">Выберите тему</option></select></div>
                <div class="form-group"><label>Текст вопроса:</label>
                <textarea name="question_text" required></textarea></div>
                <div class="form-group"><label>Правильный ответ:</label>
                <input type="text" name="correct_answer" required></div>
                <div class="form-group"><label>Неправильный ответ 1:</label>
                <input type="text" name="wrong_answer1" required></div>
                <div class="form-group"><label>Неправильный ответ 2 (необязательно):</label>
                <input type="text" name="wrong_answer2"></div>
                <div class="form-group"><label>Неправильный ответ 3 (необязательно):</label>
                <input type="text" name="wrong_answer3"></div>`;
            break;
        case 'articles':
            html = `
                <div class="form-group"><label>Тема:</label>
                <input type="text" name="topic_name" required></div>
                <div class="form-group"><label>Заголовок:</label>
                <input type="text" name="title" required></div>
                <div class="form-group"><label>Содержимое (в HTML):</label>
                <textarea name="content" rows="10" required></textarea></div>`;
            break;
    }
    document.getElementById('formFields').innerHTML = html;
    if (table === 'questions') fillTopicsSelect();
}

function closeCreateModal() {
    document.getElementById('createModal').style.display = 'none';
}

function deleteRecord(table, id) {
    if (confirm('Удалить запись?')) {
        document.getElementById('tableType').value = table;
        document.getElementById('actionType').value = 'delete';
        document.getElementById('recordId').value = id;
        document.getElementById('adminForm').submit();
    }
}

function submitForm() {
    const form = document.getElementById('createForm');
    const formData = new FormData(form);
    document.getElementById('actionType').value = formData.get('action');
    document.getElementById('tableType').value = formData.get('table');
    document.getElementById('recordId').value = formData.get('id') || '';
    for (let i = 0; i < form.elements.length; i++) {
        const element = form.elements[i];
        if (element.name && element.name !== 'action' && element.name !== 'table' && element.name !== 'id') {
            const input = document.createElement('input');
            input.type = 'hidden';
            input.name = element.name;
            input.value = element.value;
            document.getElementById('adminForm').appendChild(input);
        }
    }
    document.getElementById('adminForm').submit();
}

window.onclick = function(event) {
    const modal = document.getElementById('createModal');
    if (event.target === modal) closeCreateModal();
}