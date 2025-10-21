# ZooList - Система управления питомцами

Веб-приложение для учета питомцев и бронирований.

## Быстрый запуск

```bash
# 1. Клонируйте репозиторий
git clone <repository-url>
cd project-name

# 2. Запустите базу данных
docker run -d --name mysql-db -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=your_mysql_password \
  -e MYSQL_DATABASE=your_database_name \
  mysql:8.0

# 3. Создайте таблицы
sleep 15
docker exec -i mysql-db mysql -u root -pyour_mysql_password < init.sql

# 4. Запустите приложение
docker build -t your-app-name .
docker run -d -p 3000:8080 --name your-app-container --link mysql-db your-app-name
Приложение будет доступно: http://localhost:3000
```
## Настройка окружения
Создайте файл .env в корне проекта:

```env
DB_HOST=mysql-db
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_mysql_password
DB_NAME=your_database_name
SERVER_PORT=8080
```

## API Примеры
### Создать пользователя
```http
POST /api/register
Content-Type: application/json

{
  "nickname": "username",
  "password": "userpassword"
}
```
### Добавить питомца
```http
POST /api/pets
Content-Type: application/json

{
  "species": "Вид животного",
  "name": "Имя питомца",
  "skill_name": "Название умения"
}
```
### Получить список питомцев
```http
GET /api/pets
```
## Структура проекта
main.go - основной файл приложения
init.sql - скрипт создания таблиц БД
Dockerfile - конфигурация Docker контейнера
.env - переменные окружения (создается вручную)

## Возможные проблемы
**Ошибка подключения к БД**
  - Убедитесь, что MySQL контейнер запущен: `docker ps`
  - Проверьте пароль в `.env` файле

**Таблицы не созданы**
  - Выполните init.sql вручную: `docker exec -it mysql-db mysql -u root -p`

**Приложение не запускается**
  - Проверьте логи: `docker logs your-app-container`
