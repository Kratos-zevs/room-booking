 Room Booking Service
 Описание

REST API сервис для бронирования переговорных комнат с поддержкой расписаний и временных слотов.

Сервис позволяет управлять комнатами, задавать расписание доступности и бронировать свободные интервалы времени.

 Быстрый старт

```bash
docker-compose up --build
```

После запуска API доступно на:

http://localhost:8080

 Возможности

* Аутентификация (JWT, роли: admin / user)
* Создание и просмотр комнат
* Настройка расписания
* Генерация доступных слотов
* Бронирование слотов
* Отмена бронирования
* Просмотр своих бронирований

Технологии

* Go (Golang)
* PostgreSQL
* Docker / Docker Compose
* JWT (github.com/golang-jwt/jwt)
Аутентификация

Для упрощения используется тестовый endpoint:

```bash
POST /dummyLogin
```

Пример:

```json
{
  "role": "admin"
}
```

или

```json
{
  "role": "user"
}
```

Ответ:

```json
{
  "token": "..."
}
```
 Пример использования токена

```bash
curl -H "Authorization: Bearer <TOKEN>" http://localhost:8080/rooms/list
```
 Основные API

### Комнаты

Создать комнату (admin):

```bash
POST /rooms/create
```

Получить список:

```bash
GET /rooms/list
```

---

### Расписание

Создать расписание:

```bash
POST /schedules/create
```

Пример:

```json
{
  "room_id": 1,
  "days": [1,2,3,4,5],
  "start_time": "10:00",
  "end_time": "18:00"
}
```

---

### Слоты

Получить доступные слоты:

```bash
GET /slots?room_id=1&date=2026-03-25
```

---

### Бронирование

Создать бронь:

```bash
POST /bookings/create
```

```json
{
  "room_id": 1,
  "start_time": "10:00",
  "end_time": "10:30"
}
```

---

Отмена:

```bash
POST /bookings/cancel
```


Мои бронирования:

```bash
GET /bookings/my
```
 Пример сценария

1. Получить токен
2. Создать комнату (admin)
3. Создать расписание
4. Получить доступные слоты
5. Забронировать слот (user)
6. Проверить, что слот больше недоступен

 Структура проекта

```
cmd/app             - точка входа
internal/auth       - JWT логика
internal/db         - подключение к БД
internal/middleware - middleware авторизации
internal/repository - работа с БД
internal/service    - бизнес-логика (слоты)
```

 Ограничения

* Упрощённая аутентификация
* Нет миграций базы данных
* Минимальная валидация входных данных

 Автор

Slava Anikeev
GitHub: https://github.com/Kratos-zevs
