class AuthManager {
    static showLoginModal() {
        const modalHTML = `
            <div class="auth-modal">
                <h2>Вход в систему</h2>
                <form id="login-form">
                    <input type="text" id="login-nickname" placeholder="Никнейм" required>
                    <input type="password" id="login-password" placeholder="Пароль" required>
                    <button type="submit">Войти</button>
                </form>
                <button class="btn-cancel">Отмена</button>
            </div>
        `;
        this.showAuthModal(modalHTML, 'login-form');
    }

    static showRegisterModal() {
        const modalHTML = `
            <div class="auth-modal">
                <h2>Регистрация</h2>
                <form id="register-form">
                    <input type="text" id="register-nickname" placeholder="Никнейм" required>
                    <input type="password" id="register-password" placeholder="Пароль" required>
                    <button type="submit">Зарегистрироваться</button>
                </form>
                <button class="btn-cancel">Отмена</button>
            </div>
        `;
        this.showAuthModal(modalHTML, 'register-form');
    }

    static showAuthModal(html, formId) {
        const modalBody = document.getElementById('modal-body');
        modalBody.innerHTML = html;

        const modal = document.getElementById('pet-modal');
        modal.classList.remove('hidden');

        document.getElementById(formId).addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleAuthForm(formId);
        });

        document.querySelector('.btn-cancel').addEventListener('click', () => {
            modal.classList.add('hidden');
        });
    }

    static async handleAuthForm(formId) {
        try {
            let result;
            let credentials;

            if (formId === 'login-form') {
                credentials = {
                    nickname: document.getElementById('login-nickname').value,
                    password: document.getElementById('login-password').value
                };
                console.log('Login attempt:', credentials);
                result = await API.login(credentials);
            } else {
                const userData = {
                    nickname: document.getElementById('register-nickname').value,
                    password: document.getElementById('register-password').value
                };
                console.log('Register attempt:', userData);

                await API.register(userData);
                console.log('User registered successfully');

                credentials = userData;
                result = await API.login(credentials);
            }

            console.log('Auth result:', result);

            if (result && result.token) {
                localStorage.setItem('token', result.token);
                document.getElementById('pet-modal').classList.add('hidden');
                showMainApp();
                loadMyPets();
            } else {
                throw new Error('No token received');
            }

        } catch (error) {
            console.error('Auth error:', error);
            alert('Ошибка авторизации: ' + error.message);
        }
    }
}
