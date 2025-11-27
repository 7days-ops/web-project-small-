# Task Manager

Простой менеджер задач с аутентификацией, написанный на Go (backend) и Vue.js (frontend).

## Архитектура

- **auth-service** - сервис аутентификации (порт 8080)
- **tasks-service** - сервис управления задачами (порт 8082)
- **frontend** - Vue.js приложение (порт 3000)

## Запуск с Docker

### Предварительные требования

- Docker
- Docker Compose

### Быстрый старт

1. Запустите все сервисы:

```bash
docker-compose up --build
```

2. Откройте браузер и перейдите на `http://localhost:3000`

3. Зарегистрируйтесь и начните управлять задачами!

### Остановка сервисов

```bash
docker-compose down
```

### Остановка с удалением данных

```bash
docker-compose down -v
```

## Запуск без Docker

### Auth Service

```bash
cd auth-service
go mod download
go run main.go
```

Сервис будет доступен на `http://localhost:8080`

### Tasks Service

```bash
cd tasks-service
go mod download
go run main.go
```

Сервис будет доступен на `http://localhost:8082`

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Приложение будет доступно на `http://localhost:3000`

## API Endpoints

### Auth Service (8080)

- `POST /register` - регистрация нового пользователя
- `POST /login` - вход в систему
- `GET /verify` - проверка токена

### Tasks Service (8082)

- `GET /tasks` - получить все задачи пользователя
- `POST /tasks` - создать новую задачу
- `PUT /tasks/:id` - обновить задачу
- `DELETE /tasks/:id` - удалить задачу

## Технологии

### Backend
- Go 1.21+
- SQLite
- JWT для аутентификации

### Frontend
- Vue 3
- Vite
- Axios

## База данных

Используется SQLite для хранения данных:
- `auth.db` - пользователи
- `tasks.db` - задачи
