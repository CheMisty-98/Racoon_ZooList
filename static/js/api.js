class API {
    static baseURL = '/api';

    static async request(endpoint, options = {}) {
        const token = localStorage.getItem('token');
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...(token && { 'Authorization': `Bearer ${token}` })
            },
            ...options
        };

        try {
            const fullURL = `${this.baseURL}${endpoint}`;
            console.log('Making request to:', fullURL); // ← ДЛЯ ОТЛАДКИ

            const response = await fetch(fullURL, config);

            console.log('Response status:', response.status); // ← ДЛЯ ОТЛАДКИ

            // 🔥 ОБРАБОТКА 401 ОШИБКИ
            if (response.status === 401) {
                localStorage.removeItem('token');
                throw new Error('AUTH_REQUIRED'); // Специальный код ошибки
            }


            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            console.log('Response data:', data); // ← ДЛЯ ОТЛАДКИ
            return data;

        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }

    // Аутентификация
    static async login(credentials) {
        const response = await fetch(`${this.baseURL}/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(credentials)
        });

        if (!response.ok) {
            throw new Error(`Login failed: ${response.status}`);
        }

        const result = await response.json();
        console.log('Login response:', result);
        return result;
    }

    static async register(userData) {
        const response = await fetch(`${this.baseURL}/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(userData)
        });

        if (!response.ok) {
            throw new Error(`Registration failed: ${response.status}`);
        }

        // Регистрация возвращает текст, а не JSON
        const result = await response.text();
        console.log('Register response:', result);
        return result;
    }

    // Питомцы
    static async getMyPets() {
        return this.request('/pets/my');  // БЫЛО: /my-pets
    }

    static async getAllPets() {
        return this.request('/pets/list'); // БЫЛО: /pets
    }

    static async createPet(petData) {
        console.log('📝 Sending pet data to API:', petData); // ← ДОБАВЬ ОТЛАДКУ
        return this.request('/pets', {
            method: 'POST',
            body: JSON.stringify(petData)
        });
    }

    static async updatePet(petId, petData) {
        return this.request(`/pets/${petId}/edit`, {
            method: 'PUT',
            body: JSON.stringify(petData)
        });
    }

    static async deletePet(petId) {
        return this.request(`/pets/${petId}/delete`, {
            method: 'DELETE'
        });
    }

    // Бронирования
    static async getMyBookings() {
        return this.request('/bookings/my');
    }

    static async bookPet(petId) {
        return this.request(`/pets/${petId}/book`, {  // Теперь соответствует бэкенду
            method: 'POST',
            body: JSON.stringify({ pet_id: petId })
        });
    }

    static async cancelBooking(petId) {  // ✅ Изменено: принимает petId вместо bookingId
        return this.request(`/pets/${petId}/cancel`, {
            method: 'POST'
            // ✅ Тело запроса не нужно - petId берется из URL
        });
    }

    static async getPetQueue(petId) {
        try {
            const response = await fetch(`${this.baseURL}/pets/${petId}/queue`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            console.log('Queue response status:', response.status);

            // Читаем текст ответа, чтобы понять ошибку
            const responseText = await response.text();
            console.log('Full response text:', responseText);

            const contentType = response.headers.get('content-type');
            console.log('Content-Type:', contentType);

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}. Response: ${responseText}`);
            }

            // Проверяем, что ответ JSON
            if (!contentType || !contentType.includes('application/json')) {
                console.error('Non-JSON response:', responseText.substring(0, 200));
                throw new Error('Server returned non-JSON response');
            }

            const data = JSON.parse(responseText);
            console.log('Queue data:', data);
            return data;

        } catch (error) {
            console.error('Error in getPetQueue:', error);
            throw error;
        }
    }

    static async updateQueuePosition(petId, queueData) {
        try {
            const response = await fetch(`${this.baseURL}/pets/${petId}/queue/${queueData.direction}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: JSON.stringify({ bookingId: queueData.bookingId })
            });

            console.log('Move queue response status:', response.status);

            const responseText = await response.text();
            console.log('Move queue response text:', responseText);

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}. Response: ${responseText}`);
            }

            return JSON.parse(responseText);
        } catch (error) {
            console.error('Error in updateQueuePosition:', error);
            throw error;
        }
    }

    static async completeBooking(petId) {
        try {
            const response = await fetch(`${this.baseURL}/pets/${petId}/complete`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            console.log('Complete booking response status:', response.status);

            const responseText = await response.text();
            console.log('Complete booking response text:', responseText);

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}. Response: ${responseText}`);
            }

            return JSON.parse(responseText);
        } catch (error) {
            console.error('Error in completeBooking:', error);
            throw error;
        }
    }
}