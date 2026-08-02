# URL Shortener (в разработке)

> **Статус:** проект активно разрабатывается. Реализована базовая функциональность, ведётся работа над аналитикой, кэшированием и тестами.

---

## Описание проекта

**URL Shortener** — это сервис для сокращения длинных ссылок с возможностью сбора статистики переходов. Проект написан на Go с использованием стандартной библиотеки для HTTP, PostgreSQL в качестве хранилища и собственного in‑memory кэша для ускорения редиректов.

Основная цель — показать навыки проектирования REST API, работы с БД, асинхронной обработки данных (очередь + воркеры) и graceful shutdown.

---

## Реализованные возможности

- [x] Создание коротких ссылок с автоматической генерацией или пользовательским алиасом  
- [x] Валидация URL и алиасов  
- [x] Проверка уникальности алиаса  
- [x] Миграции базы данных (через `goose`)  
- [x] Конфигурация через переменные окружения (поддержка `.env`)  
- [x] Docker Compose для локального запуска PostgreSQL  
- [x] Редирект по короткой ссылке (`/s/{code}`) 
- [x] Асинхронная запись кликов через очередь и воркеры  
- [x] Эндпоинт статистики (`/api/v1/stats/{code}`)  
- [x] Интеграционные тесты с `httptest` и `testcontainers`  
- [x] Graceful shutdown с ожиданием завершения воркеров  
- [x] Разворачивание с Kubernetes  
- [x] Кеширование запросов (TTL Cache)  

---

## Запланированные возможности (в процессе)

- [ ] Документация API (OpenAPI/Swagger)
- [ ] Документация функций
- [ ] Доделать README
- [ ] Оформить архитектуру
---

## Технологии

- **Go** 1.23+  
- **PostgreSQL** 16  
- **pgx** — драйвер и пул соединений  
- **goose** — миграции  
- **Docker** + **Docker Compose**  
- Стандартная библиотека для HTTP (без сторонних фреймворков)  

---

## Установка и запуск

### 1. Клонирование репозитория

```bash
git clone https://github.com/yourusername/url-shortener.git
cd url-shortener
```

### 2. Переменные окружения

Создайте файл `.env` в корне проекта (пример):

```env
# База данных
DB_HOST=localhost
DB_PORT=5432
DB_NAME=shortener
DB_USER=postgres
DB_PASS=postgres
DB_SSL=disable

# Сервер
SERVER_BASE_URL=http://localhost:8000
SERVER_PORT=8000
```

### 3. Запуск базы данных через Docker Compose

```bash
docker-compose up -d
```

### 4. Применение миграций

Установите `goose` (или используйте готовый бинарник):

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Примените миграции:

```bash
goose -dir migrations postgres "postgresql://postgres:postgres@localhost:5432/shortener?sslmode=disable" up
```

### 5. Запуск приложения

```bash
go run cmd/server/main.go
```

Сервер запустится на порту, указанном в `SERVER_PORT` (по умолчанию 8000).

---

## API эндпоинты (текущая версия)

### POST /api/v1/shorten

Создание короткой ссылки.

**Тело запроса (JSON):**

```json
{
  "url": "https://example.com/very/long/url",
  "custom_alias": "mycool"   // необязательно
}
```

**Успешный ответ (201 Created):**

```json
{
  "short_url": "http://localhost:8000/s/mycool",
  "original_url": "https://example.com/very/long/url",
  "created_at": "2025-07-27T12:00:00Z"
}
```

**Возможные ошибки:**

- `400 Bad Request` — невалидный URL или алиас  
- `409 Conflict` — алиас уже занят  

---

## Структура проекта (основные пакеты)

```
cmd/
  server/            # точка входа
internal/
  config/            # загрузка конфигурации
  handler/           # HTTP-обработчики
  models/            # структуры данных
  repository/        # работа с БД
  response/          # хелперы для JSON-ответов
  service/           # бизнес-логика
migrations/          # SQL-миграции
docker-compose.yaml
```

---
