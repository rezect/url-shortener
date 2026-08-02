# URL Shortener на Go

Сервис для создания коротких ссылок с асинхронным сбором аналитики переходов (IP, User-Agent, Referer, время). Построен на слоистой архитектуре (handler → service → repository) с in-memory кэшем и очередью на каналах для батчевой записи кликов в БД.

## Стек и продемонстрированные технологии

- **Go 1.26**, стандартная библиотека `net/http` — маршрутизация через `http.ServeMux` с новым синтаксисом (`"GET /s/{alias}"`), без внешних роутеров.
- **PostgreSQL** + **pgx/v5** (`pgxpool`) — пул соединений, `pgx.Batch` для батчевой вставки кликов, транзакции (`Begin`/`WithTx`) для изоляции тестов.
- **goose** — SQL-миграции, применяются автоматически при старте (`goose.Up`), а в тестах — через `embed.FS`.
- **Конкурентность**: пул воркеров на буферизированном канале (`chan models.Click`), периодический флаш по таймеру или по размеру батча, `sync.WaitGroup` (`wg.Go`) для graceful-остановки воркеров.
- **Graceful shutdown** — обработка `SIGINT`/`SIGTERM`, поочерёдная остановка HTTP-сервера, очереди аналитики и пула БД.
- **Собственный in-memory TTL-кэш** — паттерн cache-aside с положительным и отрицательным кэшированием (несуществующие алиасы).
- **Конфигурация** — `caarlos0/env` + `godotenv`, вложенные структуры с `envPrefix`.
- **Тестирование**:
  - юнит-тесты сервисного слоя на моках (`testify/suite` + собственные mock-реализации интерфейсов);
  - интеграционные тесты репозитория на реальном PostgreSQL через `testcontainers-go`, с откатом каждого теста в транзакции;
  - HTTP-тесты хендлеров через `httptest` с моками сервиса и очереди.
- **Middleware** — логирование запросов (метод, путь, IP, длительность).
- **Docker** — multi-stage `Dockerfile`, `docker-compose.yaml` (приложение + Postgres).
- **Kubernetes** — манифесты `Deployment`/`Service`/`Ingress`/`ConfigMap`/`PVC` для приложения и БД, liveness/readiness пробы на `/health`.
- Разделение через **интерфейсы** (`Cache`, `LinkRepository`, `ClickRepository`, `Service`, `Queue`) — слои не зависят от конкретных реализаций, что упрощает мокирование.

## Архитектура

```
cmd/server/main.go        — инициализация зависимостей, миграции, graceful shutdown
internal/
  handler/                — HTTP-хендлеры, парсинг запросов, маппинг ошибок в статус-коды
  service/                — бизнес-логика: валидация, генерация алиасов, работа с кэшем
  repository/             — доступ к БД (short_links, clicks) через pgx
  cache/                  — обёртка над in-memory TTL-кэшем
  analytics/              — очередь на канале + воркеры для батчевой записи кликов
  middleware/             — логирующий middleware
  config/                 — загрузка конфигурации из env/.env
  models/                 — доменные модели (ShortLink, Click)
  response/               — хелпер для JSON-ответов
  testhelpers/            — моки и поднятие тестового Postgres-контейнера
migrations/               — SQL-миграции (goose), embed для тестов
```

Поток запроса на редирект:

1. `GET /s/{alias}` → кэш (`cache.Get`).
2. Cache hit → мгновенный редирект + событие в очередь аналитики.
3. Cache miss → запрос в БД, при находке — заполнение кэша и редирект; при отсутствии — кэширование отрицательного результата и `404`.
4. Событие клика уходит в буферизированный канал, воркеры пакетно (по размеру или по таймеру) пишут его в БД через `pgx.Batch`.

## API

### Создать короткую ссылку

```
POST /api/v1/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/url",
  "custom_alias": "mycool"   // опционально, [a-zA-Z0-9_-], 6–20 символов
}
```

**201 Created**
```json
{
  "short_url": "http://localhost:8000/s/mycool",
  "original_url": "https://example.com/very/long/url",
  "created_at": "2025-01-01T12:00:00Z"
}
```

- `400 Bad Request` — невалидный URL или алиас.
- `409 Conflict` — алиас уже занят.

### Перейти по короткой ссылке

```
GET /s/{alias}
```

- `302 Found` с заголовком `Location: <original_url>`.
- `404 Not Found` — алиас не найден.

### Получить статистику

```
GET /api/v1/stats/{short_code}
```

**200 OK**
```json
{
  "short_code": "abc123",
  "original_url": "https://example.com/very/long/url",
  "created_at": "2025-01-01T12:00:00Z",
  "total_clicks": 1234,
  "clicks_per_day": {
    "2025-01-01": 100,
    "2025-01-02": 200
  }
}
```

### Health-check

```
GET /health
```

## Запуск

### Через Docker Compose

```bash
docker-compose up --build
```

Приложение поднимется на `http://localhost:6767`, PostgreSQL — на порту `5432`. Миграции применяются автоматически при старте.

### Локально

Требуется запущенный PostgreSQL и `.env`-файл (или переменные окружения) со следующими настройками:

| Переменная                    | По умолчанию            | Описание                          |
|--------------------------------|--------------------------|------------------------------------|
| `DB_HOST`                      | `localhost`              | Хост БД                            |
| `DB_PORT`                      | `5432`                   | Порт БД                            |
| `DB_NAME`                      | `postgres`                | Имя базы                           |
| `DB_USER`                      | `postgres`                | Пользователь БД                    |
| `DB_PASSWORD`                  | `postgres`                | Пароль БД                          |
| `DB_SSL`                       | `disable`                 | Режим SSL                          |
| `SERVER_PORT`                  | `8000`                    | Порт HTTP-сервера                  |
| `SERVER_BASE_URL`              | `http://localhost:8000`   | Базовый URL для формирования коротких ссылок |
| `ANALYTICS_WORKERS`            | `5`                       | Количество воркеров аналитики      |
| `ANALYTICS_BATCH_SIZE`         | `100`                     | Размер батча для вставки кликов    |
| `ANALYTICS_FLUSH_INTERVAL_SECS`| `5`                       | Интервал принудительного флаша, сек|

```bash
go run ./cmd/server
```

### Kubernetes

В корне лежат манифесты (`backend-deployment.yaml`, `backend-service.yaml`, `backend-ingress.yaml`, `backend-config.yaml`, `postgres-deployment.yaml`, `postgres-service.yaml`, `postgres-pvc.yaml`) для развёртывания приложения и БД в кластере. Приложение ожидает `postgres-secret` с ключами `user`/`password`.

```bash
kubectl apply -f postgres-pvc.yaml -f postgres-deployment.yaml -f postgres-service.yaml
kubectl apply -f backend-config.yaml -f backend-deployment.yaml -f backend-service.yaml -f backend-ingress.yaml
```

## Тесты

```bash
go test ./...
```

- `internal/service` — юнит-тесты на моках репозиториев и кэша.
- `internal/handler` — тесты HTTP-слоя на моках сервиса и очереди (`httptest`).
- `internal/repository` — интеграционные тесты на реальном PostgreSQL, поднимаемом через `testcontainers-go`; каждый тест выполняется в отдельной транзакции с откатом.

Для интеграционных тестов репозитория нужен доступный Docker (testcontainers поднимает контейнер `postgres:16-alpine` автоматически).

## Пример использования

```bash
curl -X POST http://localhost:8000/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com/rezect/url-shortener"}'

curl -i http://localhost:8000/s/<short_code>

curl http://localhost:8000/api/v1/stats/<short_code>
```