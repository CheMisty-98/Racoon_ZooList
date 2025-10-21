console.log('ZooList loaded!');

let allPets = [];
let currentBookings = []; // ← ДОБАВИТЬ ЭТУ СТРОКУ

document.addEventListener('DOMContentLoaded', function() {
    // Проверяем авторизацию
    checkAuth();

    // Назначаем обработчики авторизации
    document.getElementById('login-btn').addEventListener('click', () => AuthManager.showLoginModal());
    document.getElementById('register-btn').addEventListener('click', () => AuthManager.showRegisterModal());
    document.getElementById('logout-btn').addEventListener('click', logout);

    // Навигация
    document.getElementById('nav-my-pets').addEventListener('click', async (e) => {
        e.preventDefault();
        await showSection('my-pets');
        setActiveNav('nav-my-pets');
    });

    document.getElementById('nav-bookings').addEventListener('click', async (e) => {
        e.preventDefault();
        await showSection('bookings');
        setActiveNav('nav-bookings');
    });

    document.getElementById('nav-all-pets').addEventListener('click', async (e) => {
        e.preventDefault();
        await showSection('all-pets');
        setActiveNav('nav-all-pets');
    });

    // Закрытие модалки
    document.querySelector('.close').addEventListener('click', function() {
        document.getElementById('pet-modal').classList.add('hidden');
    });

    // Закрытие модалки по клику снаружи
    document.getElementById('pet-modal').addEventListener('click', function(e) {
        if (e.target === this) {
            this.classList.add('hidden');
        }
    });

    document.getElementById('add-pet-btn').addEventListener('click', showAddPetModal);
});

function checkAuth() {
    const token = localStorage.getItem('token');
    if (token) {
        showMainApp();
        // loadMyPets();
    } else {
        showWelcomePage();
    }
}

function showWelcomePage() {
    document.getElementById('welcome-page').classList.remove('hidden');
    document.getElementById('main-app').classList.add('hidden');
}

function showMainApp() {
    document.getElementById('welcome-page').classList.add('hidden');
    document.getElementById('main-app').classList.remove('hidden');
}

function setActiveNav(navId) {
    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.remove('active');
    });
    document.getElementById(navId).classList.add('active');
}



// Добавьте эту функцию в app.js
function handleAuthError(error) {
    if (error.message === 'AUTH_REQUIRED') {
        showWelcomePage();
        alert('Сессия истекла. Пожалуйста, войдите снова.');
        return true;
    }
    return false;
}

// Обновите функции загрузки данных
async function loadMyPets() {
    try {
        const response = await API.getMyPets();
        console.log('My pets response:', response);
        const pets = response.myPetList || [];
        allPets = pets;
        await renderPets(pets, 'my-pets');
    } catch (error) {
        console.error('Ошибка загрузки моих питомцев:', error);
        if (!handleAuthError(error)) {
            throw error;
        }
    }
}

async function loadMyBookings() {
    try {
        const response = await API.getMyBookings();
        console.log('Bookings response:', response);
        const bookings = response.my_booking || [];
        currentBookings = bookings;
        await renderPets(bookings, 'bookings');
    } catch (error) {
        console.error('Ошибка загрузки бронирований:', error);
        if (!handleAuthError(error)) {
            throw error;
        }
    }
}

async function loadAllPets() {
    try {
        const response = await API.getAllPets();
        console.log('All pets response:', response);
        const pets = response.petList || [];
        allPets = pets;
        await renderPets(pets, 'all-pets');
    } catch (error) {
        console.error('Ошибка загрузки всех питомцев:', error);
        if (!handleAuthError(error)) {
            throw error;
        }
    }
}

// Рендер карточек
// В функцию renderPets добавляем отображение кнопки
async function renderPets(petsData, sectionType) {
    const container = document.querySelector(`#${sectionType}-content .pets-container`);

    console.log(`🎯 renderPets called for ${sectionType}:`, petsData); // ← ДОБАВЬ
    console.log(`🎯 Container found:`, container); // ← ДОБАВЬ

    container.innerHTML = '';

    if (!petsData || petsData.length === 0) {
        console.log('No pets data or empty array'); // ← ДЛЯ ОТЛАДКИ
        if (sectionType === 'my-pets') {
            container.innerHTML = `
                <div class="no-pets">
                    <p>У вас пока нет питомцев</p>
                    <p>Добавьте первого питомца!</p>
                </div>
            `;
        } else if (sectionType === 'bookings') {
            container.innerHTML = `
                <div class="no-pets">
                    <p>У вас пока нет бронирований</p>
                    <p>Забронируйте питомца из "Общего питомника"!</p>
                </div>
            `;
        } else if (sectionType === 'all-pets') {
            container.innerHTML = `
                <div class="no-pets">
                    <p>В питомнике пока нет доступных питомцев</p>
                    <p>Питомцы появятся здесь, когда другие пользователи их добавят</p>
                </div>
            `;
        }
        return;
    }

    console.log(`✅ Rendering ${petsData.length} pets for ${sectionType}`); // ← ДОБАВЬ

    // petsData.forEach((pet, index) => {
    //     console.log(`🔄 Creating card ${index} for ${sectionType}:`, pet); // ← ДОБАВЬ
    //     const card = createPetCard(pet, sectionType);
    //     container.appendChild(card);
    // });

    // СОЗДАЕМ КАРТОЧКИ АСИНХРОННО
    for (let i = 0; i < petsData.length; i++) {
        const pet = petsData[i];
        console.log(`🔄 Creating card ${i} for ${sectionType}:`, pet);
        const card = await createPetCard(pet, sectionType);
        container.appendChild(card);
    }

    console.log(`✅ Finished rendering ${sectionType}`); // ← ДОБАВЬ
}

// Функция показа модалки добавления питомца
function showAddPetModal() {
    const modal = document.getElementById('pet-modal');
    const modalBody = document.getElementById('modal-body');

    modalBody.innerHTML = `
    <div class="add-pet-modal">
        <h2>Добавить нового питомца</h2>
        <form id="add-pet-form">
            <input type="text" id="pet-name" placeholder="Имя питомца" required>
            
            <select id="pet-species" required>
                <option value="">Выберите вид</option>
                <option value="Котозомби">Котозомби</option>
                <option value="Гриб">Гриб</option>
                <option value="Волк">Волк</option>
                <option value="Динозавр">Динозавр</option>
                <option value="Ворон">Ворон</option>
                <option value="Лупоглазик">Лупоглазик</option>
                <option value="Горгулья">Горгулья</option>
                <option value="Олень">Олень</option>
                <option value="Бес">Бес</option>
                <option value="Котоангел">Котоангел</option>
                <option value="Лис">Лис</option>
                <option value="Пикси">Пикси</option>
                <option value="Подсолнух">Подсолнух</option>
                <option value="Заяц">Заяц</option>
                <option value="Осьмокрыл">Осьмокрыл</option>
                <option value="СуперОлень">СуперОлень</option>
                <option value="Люмик">Люмик</option>
                <option value="Виверна">Виверна</option>
            </select>
            
            <select id="pet-skill" required>
                <option value="">Выберите навык</option>
                <option value="Алхимия">Алхимия</option>
                <option value="Арена">Арена</option>
                <option value="Партнеры">Партнеры</option>
                <option value="Лига">Лига</option>
                <option value="Ложа">Ложа</option>
                <option value="Рыбалка">Рыбалка</option>
                <option value="Тренировки-талант">Тренировки-талант</option>
                <option value="Тренировки-навыки">Тренировки-навыки</option>
                <option value="Темная пропасть">Темная пропасть</option>
                <option value="Колодец">Колодец</option>
                <option value="Грибы">Грибы</option>
                <option value="Кристаллы">Кристаллы</option>
                <option value="Конклав">Конклав</option>
            </select>
            
            <select id="pet-skill-level" required>
                <option value="">Уровень навыка</option>
                <option value="1">I</option>
                <option value="2">II</option>
                <option value="3">III</option>
                <option value="4">IV</option>
                <option value="5">V</option>
            </select>
            
            <div class="pet-stats">
                <h3>Характеристики питомца</h3>
                <div class="stats-grid">
                    <div class="stat-input">
                        <label for="pet-loyalty">Лояльность</label>
                        <input type="number" id="pet-loyalty" min="1" max="300" value="5" required>
                    </div>
                    <div class="stat-input">
                        <label for="pet-agility">Ловкость</label>
                        <input type="number" id="pet-agility" min="1" max="300" value="5" required>
                    </div>
                    <div class="stat-input">
                        <label for="pet-stamina">Выносливость</label>
                        <input type="number" id="pet-stamina" min="1" max="300" value="5" required>
                    </div>
                    <div class="stat-input">
                        <label for="pet-instinct">Инстинкт</label>
                        <input type="number" id="pet-instinct" min="1" max="300" value="5" required>
                    </div>
                    <div class="stat-input">
                        <label for="pet-charm">Обаяние</label>
                        <input type="number" id="pet-charm" min="1" max="300" value="5" required>
                    </div>
                </div>
            </div>

            <div class="pet-options">
                <label><input type="checkbox" id="pet-ideal"> 🌟 Идеальный</label>
                <label><input type="checkbox" id="pet-mutant"> 🧬 Мутант</label>
            </div>
            <button type="submit">Добавить питомца</button>
        </form>
        <button class="btn-cancel">Отмена</button>
    </div>
`;

    modal.classList.remove('hidden');

    // Обработчик формы
    document.getElementById('add-pet-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        await createNewPet();
    });

    // Обработчик отмены
    document.querySelector('.btn-cancel').addEventListener('click', () => {
        modal.classList.add('hidden');
    });
}

// Функция создания питомца
async function createNewPet() {
    try {
        const petData = {
            Name: document.getElementById('pet-name').value,
            Species: document.getElementById('pet-species').value,
            Skill_Name: document.getElementById('pet-skill').value,
            Skill_Level: parseInt(document.getElementById('pet-skill-level').value),
            Is_Ideal: document.getElementById('pet-ideal').checked,
            Is_Mutant: document.getElementById('pet-mutant').checked,
            Stats: {
                Loyalty: parseInt(document.getElementById('pet-loyalty').value),
                Agility: parseInt(document.getElementById('pet-agility').value),
                Stamina: parseInt(document.getElementById('pet-stamina').value),
                Instinct: parseInt(document.getElementById('pet-instinct').value),
                Charm: parseInt(document.getElementById('pet-charm').value)
            }
        };

        console.log('Creating pet with correct structure:', petData);

        const result = await API.createPet(petData);
        console.log('Pet created:', result);

        document.getElementById('pet-modal').classList.add('hidden');
        await loadMyPets();

    } catch (error) {
        console.error('Error creating pet:', error);
        alert('Ошибка при создании питомца: ' + error.message);
    }
}

// В функции createPetCard для секции 'bookings' добавляем кнопку отмены
async function createPetCard(pet, sectionType) {
    const card = document.createElement('div');
    card.className = 'pet-card';

    // УНИВЕРСАЛЬНЫЕ КОНСТАНТЫ ДЛЯ ВСЕХ СЕКЦИЙ:
    const petName = pet.pet_name || pet.name || 'Неизвестно';
    const petSpecies = pet.pet_species || pet.species || 'Неизвестно';
    const skillName = pet.pet_skill_name || pet.skill_name || 'Нет навыка';
    const skillLevel = pet.pet_skill_level || pet.skill_level || '0';

    // ЖДЕМ результат функции
    const imagePath = await getPetImage(petSpecies);

    let buttons = '';
    if (sectionType === 'my-pets') {
        buttons = `
            <button class="btn-edit" onclick="editPet(${pet.id})">✏️ Редактировать</button>
            <button class="btn-delete" onclick="deletePet(${pet.id})">🗑️ Удалить</button>
        `;
    } else if (sectionType === 'all-pets') {
        buttons = `<button class="btn-book" onclick="bookPet(${pet.id})">📅 Забронировать</button>`;
    } else if (sectionType === 'bookings') {
        // ИСПРАВЛЕНО: используем pet.pet_id для отмены
        buttons = `<button class="btn-cancel-booking" onclick="cancelBooking(${pet.pet_id})">❌ Отменить бронь</button>`;
    }

    // ИСПРАВЛЕННЫЙ HTML С УНИВЕРСАЛЬНЫМИ ПОЛЯМИ:
    card.innerHTML = `
        <img src="${imagePath}" alt="${petName}" onerror="this.src='/static/images/pets/default.svg'">
        <h3>${petName}</h3>
        <p>Вид: ${petSpecies}</p>
        <p>Навык: ${skillName} (уровень ${skillLevel})</p>
        ${pet.booking_position ? `<p>Позиция в очереди: ${pet.booking_position}</p>` : ''}
        ${pet.is_ideal ? '<span class="badge ideal">🌟 Идеальный</span>' : ''}
        ${pet.is_mutant ? '<span class="badge mutant">🧬 Мутант</span>' : ''}
        <div class="pet-actions">
            <button class="btn-details" onclick="showPetDetails(${pet.pet_id || pet.id}, '${sectionType}')">
                👀 Подробнее
            </button>
            ${buttons}
        </div>
    `;

    return card;
}

async function getPetImage(species) {
    const formats = ['webp', 'svg', 'png', 'jpg'];
    const fileName = species.toLowerCase().replace(/\s+/g, '_');

    for (let format of formats) {
        const path = `/static/images/pets/${fileName}.${format}`;
        // Можно добавить проверку существования файла
        if (await imageExists(path)) return path;
    }
    return '/static/images/pets/default.svg';
}

function imageExists(url) {
    return new Promise((resolve) => {
        const img = new Image();
        img.onload = () => resolve(true);
        img.onerror = () => resolve(false);
        img.src = url;
    });
}

// Функция отмены бронирования
async function cancelBooking(petId) {
    if (!confirm('Вы уверены, что хотите отменить бронирование?')) {
        return;
    }

    try {
        await API.cancelBooking(petId);
        console.log('Booking cancelled for pet:', petId);
        alert('Бронирование отменено!');
        await loadMyBookings();
        await loadAllPets();
    } catch (error) {
        console.error('Error cancelling booking:', error);
        alert('Ошибка при отмене бронирования: ' + error.message);
    }
}

// Модальное окно питомца
async function showPetDetails(petId, sectionType) {
    let pet;

    // ПОИСК ДЛЯ ВСЕХ СЕКЦИЙ:
    if (sectionType === 'bookings') {
        pet = currentBookings.find(b => b.pet_id === petId);
    } else {
        // Для all-pets и my-pets ищем в allPets
        pet = allPets.find(p => p.id === petId);
    }

    if (!pet) {
        console.error('Pet not found:', petId, 'in section:', sectionType);
        alert('Данные питомца не найдены');
        return;
    }

    const modal = document.getElementById('pet-modal');
    const modalBody = document.getElementById('modal-body');

    let actionButtons = '';
    if (sectionType === 'my-pets') {
        actionButtons = `
            <button class="btn-edit" onclick="editPet(${pet.id})">✏️ Редактировать питомца</button>
            <button class="btn-book" onclick="manageQueue(${pet.id})">📋 Управлять очередью</button>
        `;
    } else if (sectionType === 'all-pets') {
        actionButtons = `<button class="btn-book" onclick="bookPet(${pet.id})">📅 Забронировать</button>`;
    } else if (sectionType === 'bookings') {
        actionButtons = `<button class="btn-cancel-booking" onclick="cancelBooking(${pet.pet_id})">❌ Отменить бронирование</button>`;
    }

    // УНИВЕРСАЛЬНЫЕ ПОЛЯ ДЛЯ ОТОБРАЖЕНИЯ:
    const petName = pet.pet_name || pet.name || 'Неизвестно';
    const petSpecies = pet.pet_species || pet.species || 'Неизвестно';
    const skillName = pet.pet_skill_name || pet.skill_name || 'Неизвестно';
    const skillLevel = pet.pet_skill_level || pet.skill_level || '?';

    modalBody.innerHTML = `
        <div class="pet-details">
            <h2>${petName}</h2>
            <img src="https://via.placeholder.com/300x200/667eea/ffffff?text=${encodeURIComponent(petSpecies)}" alt="${petName}" style="width: 200px; border-radius: 10px;">
            
            <div class="details-info">
                <p><strong>Вид:</strong> ${petSpecies}</p>
                <p><strong>Навык:</strong> ${skillName} (уровень ${skillLevel})</p>
                
                ${pet.booking_position ? `<p><strong>Позиция в очереди:</strong> ${pet.booking_position}</p>` : ''}
                ${pet.estimated_wait_time ? `<p><strong>Примерное время ожидания:</strong> ${pet.estimated_wait_time} часов</p>` : ''}
                
                <!-- Характеристики -->
                ${(pet.stats || pet.loyalty) ? `
                <div class="pet-stats">
                    <h3>Характеристики:</h3>
                    <ul>
                        ${pet.stats ? `
                            <li>Лояльность: ${pet.stats.loyalty || 0}</li>
                            <li>Ловкость: ${pet.stats.agility || 0}</li>
                            <li>Выносливость: ${pet.stats.stamina || 0}</li>
                            <li>Инстинкт: ${pet.stats.instinct || 0}</li>
                            <li>Обаяние: ${pet.stats.charm || 0}</li>
                        ` : `
                            <li>Лояльность: ${pet.loyalty || 0}</li>
                            <li>Ловкость: ${pet.agility || 0}</li>
                            <li>Выносливость: ${pet.stamina || 0}</li>
                            <li>Инстинкт: ${pet.instinct || 0}</li>
                            <li>Обаяние: ${pet.charm || 0}</li>
                        `}
                    </ul>
                </div>
                ` : ''}
                
                <!-- Секция очереди для ВСЕХ питомцев -->
                <div class="queue-section">
                    <h3>Очередь бронирований</h3>
                    <div id="queue-content">
                        <p>Загрузка очереди...</p>
                    </div>
                </div>
            </div>
            
            <div class="action-buttons">
                ${actionButtons}
                <button class="btn-cancel" onclick="document.getElementById('pet-modal').classList.add('hidden')">Закрыть</button>
            </div>
        </div>
    `;

    modal.classList.remove('hidden');

    // Загружаем очередь только для моих питомцев
        try {
            await loadQueueForPet(petId);
        } catch (error) {
            console.error('Ошибка загрузки очереди:', error);
            document.getElementById('queue-content').innerHTML =
                '<p>Ошибка загрузки очереди</p>';
        }
}

// Функция загрузки очереди для питомца
async function loadQueueForPet(petId) {
    try {
        const queue = await API.getPetQueue(petId);
        const queueContent = document.getElementById('queue-content');

        if (!queue || queue.length === 0) {
            queueContent.innerHTML = '<p>Очередь пуста</p>';
        } else {
            queueContent.innerHTML = `
                <div class="queue-list">
                    ${queue.map((booking, index) => `
                        <div class="queue-item ${index === 0 ? 'current' : ''}">
                            <span class="queue-position">${index + 1}.</span>
                            <span class="queue-user">${booking.nickname}</span>
                            <span class="wait-time">${formatWaitTime(booking.ready_in)}</span>
                            ${index === 0 ? '<span class="badge current-badge">Текущий</span>' : ''}
                        </div>
                    `).join('')}
                </div>
            `;
        }
    } catch (error) {
        console.error('Error loading queue:', error);
        document.getElementById('queue-content').innerHTML =
            '<p>Ошибка загрузки очереди: очередь недоступна</p>';
    }
}

async function showSection(sectionName) {
    // Скрываем все секции
    document.querySelectorAll('.content-section').forEach(section => {
        section.classList.add('hidden');
        section.classList.remove('active');
    });

    // Показываем нужную секцию
    const targetSection = document.getElementById(`${sectionName}-content`);
    targetSection.classList.remove('hidden');
    targetSection.classList.add('active');

    // Показываем индикатор загрузки
    const container = targetSection.querySelector('.pets-container');
    container.innerHTML = '<div class="loading">Загрузка...</div>';

    try {
        await loadSectionData(sectionName);
    } catch (error) {
        console.error(`Ошибка загрузки секции ${sectionName}:`, error);
        container.innerHTML = '<div class="error">Ошибка загрузки данных</div>';
        alert(`Ошибка загрузки ${sectionName}: ${error.message}`); // ← alert здесь
    }
}

async function loadSectionData(sectionName) {
    switch(sectionName) {
        case 'my-pets':
            await loadMyPets();
            break;
        case 'bookings':
            await loadMyBookings();
            break;
        case 'all-pets':
            await loadAllPets();
            break;
    }
}

function logout() {
    localStorage.removeItem('token');
    showWelcomePage();
}

async function editPet(petId) {
    const pet = allPets.find(p => p.id === petId);
    if (!pet) return;

    const modal = document.getElementById('pet-modal');
    const modalBody = document.getElementById('modal-body');

    modalBody.innerHTML = `
        <div class="add-pet-modal">
            <h2>Редактировать питомца</h2>
            <form id="edit-pet-form">
                <input type="text" id="edit-pet-name" value="${pet.name}" placeholder="Имя питомца" required>
                
                <select id="edit-pet-species" required>
                    <option value="">Выберите вид</option>
                    <option value="Котозомби" ${pet.species === 'Котозомби' ? 'selected' : ''}>Котозомби</option>
                    <option value="Гриб" ${pet.species === 'Гриб' ? 'selected' : ''}>Гриб</option>
                    <option value="Волк" ${pet.species === 'Волк' ? 'selected' : ''}>Волк</option>
                    <option value="Динозавр" ${pet.species === 'Динозавр' ? 'selected' : ''}>Динозавр</option>
                    <option value="Ворон" ${pet.species === 'Ворон' ? 'selected' : ''}>Ворон</option>
                    <option value="Лупоглазик" ${pet.species === 'Лупоглазик' ? 'selected' : ''}>Лупоглазик</option>
                    <option value="Горгулья" ${pet.species === 'Горгулья' ? 'selected' : ''}>Горгулья</option>
                    <option value="Олень" ${pet.species === 'Олень' ? 'selected' : ''}>Олень</option>
                    <option value="Бес" ${pet.species === 'Бес' ? 'selected' : ''}>Бес</option>
                    <option value="Котоангел" ${pet.species === 'Котоангел' ? 'selected' : ''}>Котоангел</option>
                    <option value="Лис" ${pet.species === 'Лис' ? 'selected' : ''}>Лис</option>
                    <option value="Пикси" ${pet.species === 'Пикси' ? 'selected' : ''}>Пикси</option>
                    <option value="Подсолнух" ${pet.species === 'Подсолнух' ? 'selected' : ''}>Подсолнух</option>
                    <option value="Заяц" ${pet.species === 'Заяц' ? 'selected' : ''}>Заяц</option>
                    <option value="Осьмокрыл" ${pet.species === 'Осьмокрыл' ? 'selected' : ''}>Осьмокрыл</option>
                    <option value="СуперОлень" ${pet.species === 'СуперОлень' ? 'selected' : ''}>СуперОлень</option>
                    <option value="Люмик" ${pet.species === 'Люмик' ? 'selected' : ''}>Люмик</option>
                    <option value="Виверна" ${pet.species === 'Виверна' ? 'selected' : ''}>Виверна</option>
                </select>
                
                <select id="edit-pet-skill" required>
                    <option value="">Выберите навык</option>
                    <option value="Алхимия" ${pet.skill_name === 'Алхимия' ? 'selected' : ''}>Алхимия</option>
                    <option value="Арена" ${pet.skill_name === 'Арена' ? 'selected' : ''}>Арена</option>
                    <option value="Партнеры" ${pet.skill_name === 'Партнеры' ? 'selected' : ''}>Партнеры</option>
                    <option value="Лига" ${pet.skill_name === 'Лига' ? 'selected' : ''}>Лига</option>
                    <option value="Ложа" ${pet.skill_name === 'Ложа' ? 'selected' : ''}>Ложа</option>
                    <option value="Рыбалка" ${pet.skill_name === 'Рыбалка' ? 'selected' : ''}>Рыбалка</option>
                    <option value="Тренировки-талант" ${pet.skill_name === 'Тренировки-талант' ? 'selected' : ''}>Тренировки-талант</option>
                    <option value="Тренировки-навыки" ${pet.skill_name === 'Тренировки-навыки' ? 'selected' : ''}>Тренировки-навыки</option>
                    <option value="Темная пропасть" ${pet.skill_name === 'Темная пропасть' ? 'selected' : ''}>Темная пропасть</option>
                    <option value="Колодец" ${pet.skill_name === 'Колодец' ? 'selected' : ''}>Колодец</option>
                    <option value="Грибы" ${pet.skill_name === 'Грибы' ? 'selected' : ''}>Грибы</option>
                    <option value="Кристаллы" ${pet.skill_name === 'Кристаллы' ? 'selected' : ''}>Кристаллы</option>
                    <option value="Конклав" ${pet.skill_name === 'Конклав' ? 'selected' : ''}>Конклав</option>
                </select>
                
                <select id="edit-pet-skill-level" required>
                    <option value="">Уровень навыка</option>
                    <option value="1" ${pet.skill_level === 1 ? 'selected' : ''}>I</option>
                    <option value="2" ${pet.skill_level === 2 ? 'selected' : ''}>II</option>
                    <option value="3" ${pet.skill_level === 3 ? 'selected' : ''}>III</option>
                    <option value="4" ${pet.skill_level === 4 ? 'selected' : ''}>IV</option>
                    <option value="5" ${pet.skill_level === 5 ? 'selected' : ''}>V</option>
                </select>
                
                <div class="pet-stats">
                    <h3>Характеристики питомца</h3>
                    <div class="stats-grid">
                        <div class="stat-input">
                            <label for="edit-pet-loyalty">Лояльность</label>
                            <input type="number" id="edit-pet-loyalty" value="${pet.loyalty || 5}" min="1" max="300" required>
                        </div>
                        <div class="stat-input">
                            <label for="edit-pet-agility">Ловкость</label>
                            <input type="number" id="edit-pet-agility" value="${pet.agility || 5}" min="1" max="300" required>
                        </div>
                        <div class="stat-input">
                            <label for="edit-pet-stamina">Выносливость</label>
                            <input type="number" id="edit-pet-stamina" value="${pet.stamina || 5}" min="1" max="300" required>
                        </div>
                        <div class="stat-input">
                            <label for="edit-pet-instinct">Инстинкт</label>
                            <input type="number" id="edit-pet-instinct" value="${pet.instinct || 5}" min="1" max="300" required>
                        </div>
                        <div class="stat-input">
                            <label for="edit-pet-charm">Обаяние</label>
                            <input type="number" id="edit-pet-charm" value="${pet.charm || 5}" min="1" max="300" required>
                        </div>
                    </div>
                </div>

                <div class="pet-options">
                    <label><input type="checkbox" id="edit-pet-ideal" ${pet.is_ideal ? 'checked' : ''}> 🌟 Идеальный</label>
                    <label><input type="checkbox" id="edit-pet-mutant" ${pet.is_mutant ? 'checked' : ''}> 🧬 Мутант</label>
                </div>
                <button type="submit">Сохранить изменения</button>
            </form>
            <button class="btn-cancel">Отмена</button>
        </div>
    `;

    modal.classList.remove('hidden');

    document.getElementById('edit-pet-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        await updatePet(petId);
    });

    document.querySelector('.btn-cancel').addEventListener('click', () => {
        modal.classList.add('hidden');
    });
}

async function updatePet(petId) {
    try {
        const petData = {
            Name: document.getElementById('edit-pet-name').value,
            Species: document.getElementById('edit-pet-species').value,
            Skill_Name: document.getElementById('edit-pet-skill').value,
            Skill_Level: parseInt(document.getElementById('edit-pet-skill-level').value),
            Is_Ideal: document.getElementById('edit-pet-ideal').checked,
            Is_Mutant: document.getElementById('edit-pet-mutant').checked,
            Stats: {
                Loyalty: parseInt(document.getElementById('edit-pet-loyalty').value),
                Agility: parseInt(document.getElementById('edit-pet-agility').value),
                Stamina: parseInt(document.getElementById('edit-pet-stamina').value),
                Instinct: parseInt(document.getElementById('edit-pet-instinct').value),
                Charm: parseInt(document.getElementById('edit-pet-charm').value)
            }
        };

        await API.updatePet(petId, petData);
        document.getElementById('pet-modal').classList.add('hidden');
        await loadMyPets();

    } catch (error) {
        console.error('Error updating pet:', error);
        alert('Ошибка при обновлении питомца: ' + error.message);
    }
}

async function deletePet(petId) {
    if (!confirm('Вы уверены, что хотите удалить питомца? Это действие нельзя отменить.')) {
        return;
    }

    try {
        console.log('Deleting pet:', petId);
        await API.deletePet(petId);
        console.log('Pet deleted successfully:', petId);

        // Обновляем интерфейс
        await loadMyPets();

        // Закрываем модалку если открыта
        document.getElementById('pet-modal').classList.add('hidden');

        alert('Питомец успешно удален!');

    } catch (error) {
        console.error('Error deleting pet:', error);
        alert('Ошибка при удалении питомца: ' + error.message);
    }
}

async function bookPet(petId) {
    try {
        const result = await API.bookPet(petId);
        console.log('Pet booked:', result);
        alert('Питомец успешно забронирован!');
        await loadAllPets(); // Обновляем список (питомец исчезнет из доступных)
    } catch (error) {
        console.error('Error booking pet:', error);
        alert('Ошибка при бронировании: ' + error.message);
    }
}

async function manageQueue(petId) {
    try {
        const queue = await API.getPetQueue(petId);
        const pet = allPets.find(p => p.id === petId);

        const modal = document.getElementById('pet-modal');
        const modalBody = document.getElementById('modal-body');

        modalBody.innerHTML = `
            <div class="queue-management">
                <h2>Очередь бронирований: ${pet.name}</h2>
                
                ${queue.length > 0 && queue[0].ready_in === 0 ? `
                    <div class="completion-section">
                        <p>🐾 Питомец готов к использованию!</p>
                        <button class="btn-complete" onclick="completeBooking(${petId})">
                            ✅ Подтвердить завершение брони
                        </button>
                    </div>
                ` : ''}
                
                <div class="queue-management-list">
                    ${queue.length === 0 ?
            '<p class="empty-queue">Очередь пуста</p>' :
            queue.map((booking, index) => `
                            <div class="management-item ${index === 0 ? 'current' : ''}" 
                                 onclick="selectQueueItem(${index})"
                                 id="queue-item-${index}">
                                <div class="item-info">
                                    <span class="item-position">${index + 1}.</span>
                                    <span class="item-user">${booking.nickname}</span>
                                    <span class="item-time">${formatWaitTime(booking.ready_in)}</span>
                                </div>
                                <div class="item-actions">
                                    <button class="btn-up" onclick="event.stopPropagation(); moveInQueue(${petId}, ${booking.id}, 'up')" 
                                        ${index === 0 ? 'disabled' : ''}>↑</button>
                                    <button class="btn-down" onclick="event.stopPropagation(); moveInQueue(${petId}, ${booking.id}, 'down')" 
                                        ${index === queue.length - 1 ? 'disabled' : ''}>↓</button>
                                </div>
                            </div>
                        `).join('')
        }
                </div>
                
                <button class="btn-cancel">Закрыть</button>
            </div>
        `;

        modal.classList.remove('hidden');

        document.querySelector('.btn-cancel').addEventListener('click', () => {
            modal.classList.add('hidden');
        });

    } catch (error) {
        console.error('Error loading queue:', error);
        alert('Ошибка при загрузке очереди: ' + error.message);
    }
}

// Функция для выделения выбранного элемента
function selectQueueItem(index) {
    // Снимаем выделение со всех элементов
    document.querySelectorAll('.management-item').forEach(item => {
        item.classList.remove('selected');
    });
    // Выделяем выбранный элемент
    document.getElementById(`queue-item-${index}`).classList.add('selected');
}

async function completeBooking(petId) {
    try {
        await API.completeBooking(petId);
        alert('Бронирование завершено! Питомец снова доступен.');
        await manageQueue(petId); // Обновляем очередь
    } catch (error) {
        console.error('Error completing booking:', error);
        alert('Ошибка при завершении брони: ' + error.message);
    }
}

async function moveInQueue(petId, bookingId, direction) {
    try {
        await API.updateQueuePosition(petId, { bookingId, direction });
        await manageQueue(petId); // Перезагружаем очередь
    } catch (error) {
        console.error('Error moving in queue:', error);
        alert('Ошибка при перемещении в очереди: ' + error.message);
    }
}

function formatWaitTime(hours) {
    if (hours === 0) return '✅ Готов сейчас';

    const days = Math.floor(hours / 24);
    const remainingHours = hours % 24;

    if (days > 0) {
        return `⏳ Готов через: ${hours}ч (~${days} ${getDayWord(days)})`;
    } else {
        return `⏳ Готов через: ${hours}ч`;
    }
}

function getDayWord(days) {
    if (days === 1) return 'день';
    if (days >= 2 && days <= 4) return 'дня';
    return 'дней';
}