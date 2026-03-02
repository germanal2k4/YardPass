# YardPass Backend

Backend система для управления гостевых пропусков во двор/ЖК с Telegram-ботом.

## Описание

YardPass позволяет жителям создавать временные гостевые пропуска через Telegram-бота. Охранники сканируют QR-коды через веб-панель, система валидирует пропуск и записывает события в журнал.

## Технологии

- **Go 1.25+**
- **Gin** - HTTP framework
- **PostgreSQL 15+** - основная БД
- **Redis** - кеш, rate limiting, состояния бота
- **JWT** - аутентификация для веб-пользователей
- **Telegram Bot API** - бот для жителей
- **goose** (pressly/goose) - версионирование и применение миграций БД
- **sqlc** - генерация типобезопасного Go-кода из SQL-запросов

## Структура проекта

```
/cmd
  /api      - точка входа для HTTP API
  /bot      - точка входа для Telegram бота
  /migrate  - CLI для управления миграциями (goose)
/internal
  /auth     - JWT аутентификация
  /config   - конфигурация
  /db       - сгенерированный sqlc-код (модели, запросы, интерфейс Querier)
  /db/queries - .sql файлы с аннотированными SQL-запросами для sqlc
  /domain   - доменные модели и интерфейсы репозиториев
  /http     - HTTP handlers, middleware, роуты
  /observability - логирование, трейсинг, метрики
  /qr       - генерация QR кодов
  /redis    - Redis клиент
  /repo     - реализации репозиториев (обёртки над sqlc)
  /service  - бизнес-логика
  /telegram - Telegram бот
/migrations - SQL миграции (формат goose: up + down в одном файле)
```

## Установка и запуск

### Требования

- Go 1.25+
- PostgreSQL 15+
- Redis 6+
- sqlc (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

### Настройка

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd YardPass
```

2. Установите зависимости:
```bash
go mod download
```

3. Создайте файл `.env` на основе `.env.example`:
```bash
cp .env.example .env
```

4. Настройте переменные окружения в `.env`:
```env
DATABASE_URL=postgres://user:password@localhost:5432/yardpass?sslmode=disable
REDIS_URL=redis://localhost:6379/0
JWT_SECRET=your-secret-key
TELEGRAM_BOT_TOKEN=your-bot-token
SERVICE_TOKEN=your-service-token
```

5. Создайте базу данных и выполните миграции:
```bash
createdb yardpass

# Применить все миграции (goose)
go run ./cmd/migrate up

# Посмотреть статус
go run ./cmd/migrate status
```

Тестовые пользователи создаются автоматически миграцией 005:
- `superadmin` / `admin123` (superuser)
- `admin` / `admin123` (admin)
- `guard` / `guard123` (guard)

### Запуск

#### API сервер

```bash
make run-api
# или
go run ./cmd/api
```

API будет доступен на `http://localhost:8080`

#### Telegram бот

```bash
make run-bot
# или
go run ./cmd/bot
```

Бот работает в режиме polling или webhook (настраивается через `TELEGRAM_WEBHOOK_URL`)

### Сборка

```bash
make build
```

Бинарные файлы будут в `bin/api` и `bin/bot`

## API Endpoints

### Аутентификация

- `POST /auth/login` - вход (получить JWT токены)
- `POST /auth/refresh` - обновить access token
- `GET /api/v1/me` - информация о текущем пользователе

### Пропуска

- `POST /api/v1/passes` - создать пропуск (требует аутентификации)
- `GET /api/v1/passes/:id` - получить пропуск по ID
- `POST /api/v1/passes/:id/revoke` - отозвать пропуск
- `POST /api/v1/passes/validate` - валидировать QR код (для охранников)
- `GET /api/v1/passes/active` - список активных пропусков

### Правила (только для админов)

- `GET /api/v1/rules?building_id=1` - получить правила для здания
- `PUT /api/v1/rules?building_id=1` - обновить правила

### Service API (для бота)

- `POST /service/v1/passes` - создать пропуск (service token)
- `POST /service/v1/passes/:id/revoke` - отозвать пропуск
- `GET /service/v1/passes/active?apartment_id=1` - активные пропуска

## Формат ошибок

Все ошибки возвращаются в едином формате:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message"
  }
}
```

### Коды ошибок

- `PASS_NOT_FOUND` - пропуск не найден
- `PASS_EXPIRED` - пропуск истек
- `PASS_REVOKED` - пропуск отозван
- `PASS_NOT_YET_VALID` - пропуск еще не действителен
- `QUIET_HOURS` - действие запрещено в тихие часы
- `RATE_LIMIT_EXCEEDED` - превышен лимит запросов
- `INVALID_CREDENTIALS` - неверные учетные данные
- `INVALID_TOKEN` - неверный или истекший токен
- `INSUFFICIENT_PERMISSIONS` - недостаточно прав

## Telegram бот

### Команды

- `/start` - начать работу с ботом

### Флоу создания пропуска

1. Нажать "Выдать пропуск гостю"
2. Ввести номер автомобиля
3. Выбрать срок действия (1 час, 2 часа, 4 часа, или до времени)
4. Ввести имя гостя (опционально)
5. Получить QR код

### Флоу просмотра пропусков

1. Нажать "Мои активные пропуска"
2. Просмотреть список активных пропусков

## Тестирование

```bash
make test
# или
go test -v ./...
```

## Линтинг

```bash
make lint
```

Требуется установленный `golangci-lint`.

## Конфигурация

Все настройки через переменные окружения (см. `.env.example`):

- `SERVER_HOST`, `SERVER_PORT` - адрес API сервера
- `DATABASE_URL` - строка подключения к PostgreSQL
- `REDIS_URL` - строка подключения к Redis
- `JWT_SECRET` - секретный ключ для JWT
- `JWT_ACCESS_TTL`, `JWT_REFRESH_TTL` - время жизни токенов
- `TELEGRAM_BOT_TOKEN` - токен Telegram бота
- `TELEGRAM_WEBHOOK_URL` - URL для webhook (опционально)
- `SERVICE_TOKEN` - токен для service API
- `RATE_LIMIT_*` - настройки rate limiting
- `LOG_LEVEL`, `LOG_FORMAT` - настройки логирования

## Безопасность

- Все пароли хешируются с помощью bcrypt
- JWT токены с коротким временем жизни
- Rate limiting на критичных endpoints
- Валидация всех входных данных
- SQL injection защита через параметризованные запросы

## Разработка

### Миграции БД

Миграции управляются через [goose](https://github.com/pressly/goose).
Каждый файл в `migrations/` содержит секции `-- +goose Up` и `-- +goose Down`.

```bash
# Создать новую миграцию
go run ./cmd/migrate create add_new_feature sql

# Применить все миграции
go run ./cmd/migrate up

# Откатить последнюю
go run ./cmd/migrate down

# Статус
go run ./cmd/migrate status
```

### Изменение SQL-запросов или схемы

1. Измените/добавьте миграцию в `migrations/`
2. Обновите SQL-запросы в `internal/db/queries/*.sql`
3. Перегенерируйте код: `sqlc generate`
4. При необходимости обновите маппинг в `internal/repo/`
5. Закоммитьте сгенерированный код вместе с изменениями

CI автоматически проверяет, что `sqlc generate && git diff --exit-code` проходит.

### Добавление нового endpoint

1. Добавьте handler в `internal/api/handlers/`
2. Зарегистрируйте роут в `internal/api/router.go`
3. Добавьте тесты

### Добавление новой бизнес-логики

1. Добавьте метод в соответствующий сервис в `internal/service/`
2. Используйте репозитории для доступа к данным
3. Добавьте валидацию и обработку ошибок

## Лицензия

[Укажите лицензию]

