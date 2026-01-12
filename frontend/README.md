# YardPass Frontend

Веб-панель для системы управления гостевыми пропусками YardPass.

## Описание

Frontend приложение предоставляет интерфейсы для двух ролей:

- **Охрана** - сканирование QR-кодов пропусков, просмотр результатов проверки, журнал событий
- **Администратор** - настройка правил (тихие часы, лимиты), просмотр отчетов и статистики

## Технологии

- **React 18** + TypeScript
- **Vite** - сборщик
- **Material-UI (MUI)** - UI компоненты
- **React Router** - роутинг
- **TanStack Query** - управление серверным состоянием
- **React Hook Form** + Zod - формы и валидация
- **Axios** - HTTP клиент
- **date-fns** - работа с датами

## Требования

- Node.js 18+ 
- npm / pnpm / yarn
- Backend API (должен быть запущен на `http://localhost:8080`)

## Установка и запуск

### 1. Установка зависимостей

```bash
cd frontend/app
npm install
# или
pnpm install
# или
yarn install
```

### 2. Настройка переменных окружения

Создайте файл `.env` на основе `.env.example`:

```bash
cp .env.example .env
```

Отредактируйте `.env` при необходимости:

```env
VITE_API_BASE_URL=http://localhost:8080
```

### 3. Запуск в режиме разработки

```bash
npm run dev
# или
pnpm dev
# или
yarn dev
```

Приложение будет доступно на `http://localhost:3000`

Vite автоматически проксирует запросы `/api`, `/auth`, `/health` на backend (настроено в `vite.config.ts`).

### 4. Сборка для production

```bash
npm run build
# или
pnpm build
# или
yarn build
```

Результат сборки будет в папке `dist/`.

### 5. Предпросмотр production сборки

```bash
npm run preview
# или
pnpm preview
# или
yarn preview
```

## Запуск с Docker

### Сборка образа

```bash
cd frontend/app
docker build -t yardpass-frontend .
```

### Запуск контейнера

```bash
docker run -d -p 3000:80 --name yardpass-frontend yardpass-frontend
```

**Важно:** При запуске в Docker убедитесь, что в `nginx.conf` указан правильный адрес backend-сервиса. По умолчанию используется `http://backend:8080`, что подходит для docker-compose.

### Docker Compose (вместе с backend)

Создайте `docker-compose.yml` в корне проекта:

```yaml
version: '3.8'

services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:password@postgres:5432/yardpass
      - REDIS_URL=redis://redis:6379/0
      - JWT_SECRET=your-secret-key
    depends_on:
      - postgres
      - redis

  frontend:
    build: ./frontend/app
    ports:
      - "3000:80"
    depends_on:
      - backend

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=yardpass
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

Запуск:

```bash
docker-compose up -d
```

## Структура проекта

```
frontend/app/
├── src/
│   ├── app/              # Инициализация приложения, роутинг
│   │   ├── App.tsx
│   │   └── router.tsx
│   ├── pages/            # Страницы
│   │   ├── LoginPage.tsx
│   │   ├── SecurityPage.tsx
│   │   ├── AdminPage.tsx
│   │   ├── AdminRulesPage.tsx
│   │   ├── AdminReportsPage.tsx
│   │   └── ForbiddenPage.tsx
│   ├── features/         # Фичи и логика по доменам
│   │   ├── auth/
│   │   │   ├── AuthProvider.tsx
│   │   │   └── useAuth.ts
│   │   └── security/
│   │       ├── PassDetailsCard.tsx
│   │       └── EventsLog.tsx
│   ├── shared/           # Общий код
│   │   ├── api/          # HTTP клиент и API методы
│   │   │   ├── client.ts
│   │   │   ├── auth.ts
│   │   │   ├── passes.ts
│   │   │   └── rules.ts
│   │   ├── config/       # Конфигурация
│   │   │   ├── env.ts
│   │   │   └── constants.ts
│   │   ├── types/        # TypeScript типы
│   │   │   └── api.ts
│   │   └── ui/           # UI компоненты
│   │       ├── Layout.tsx
│   │       └── theme.ts
│   └── main.tsx          # Точка входа
├── index.html
├── vite.config.ts
├── tsconfig.json
├── package.json
├── Dockerfile
├── nginx.conf
└── README.md
```

## Функционал

### Роль "Охрана" (`/security`)

- **Сканирование QR-кода**: фокусируемое поле ввода, работает с HID-сканером (вводит строку + Enter)
- **Результат проверки**: карточка с деталями пропуска (статус, номер авто, квартира, срок действия)
- **Звуковой фидбэк**: звуковой сигнал при успешной/неуспешной проверке
- **Журнал событий**: таблица последних сканирований (требует добавления endpoint в backend)

### Роль "Администратор" (`/admin`)

- **Главная панель**: навигация по разделам (Правила, Отчеты)
- **Настройка правил** (`/admin/rules`):
  - Тихие часы (начало и конец)
  - Лимит пропусков в день на квартиру
  - Максимальный срок действия пропуска
- **Отчеты** (`/admin/reports`):
  - Placeholder с описанием требуемых backend endpoints
  - Планируется: журнал событий, статистика, экспорт в Excel, информация о парковке

### Аутентификация

- JWT токены (access + refresh)
- Автоматическое обновление access token при истечении
- Защита роутов по ролям (guard / admin)
- Logout с очисткой токенов

## API интеграция

Приложение интегрируется с backend через REST API. Подробности API см. в `backend/docs/openapi.yaml`.

### Используемые endpoints:

- `POST /auth/login` - вход
- `POST /auth/refresh` - обновление токена
- `GET /api/v1/me` - информация о пользователе
- `POST /api/v1/passes/validate` - валидация QR-кода
- `GET /api/v1/rules?building_id=1` - получение правил
- `PUT /api/v1/rules?building_id=1` - обновление правил

### Отсутствующие endpoints (требуется добавить в backend):

- `GET /api/v1/scan-events` - журнал событий сканирования
- `GET /api/v1/reports/statistics` - статистика
- `GET /api/v1/reports/export` - экспорт в Excel
- `GET /api/v1/parking/occupancy` - загруженность парковки
- `GET /api/v1/parking/vehicles` - список автомобилей

## Разработка

### Линтинг

```bash
npm run lint
```

### Проверка типов

```bash
npx tsc --noEmit
```

### Форматирование (если настроен Prettier)

```bash
npx prettier --write src/
```

## Тестовые пользователи

Для тестирования создайте пользователей в backend:

```sql
-- Охранник
INSERT INTO users (username, password_hash, role, status) 
VALUES ('guard1', '$2a$10$...', 'guard', 'active');

-- Администратор
INSERT INTO users (username, password_hash, role, status) 
VALUES ('admin', '$2a$10$...', 'admin', 'active');
```

Используйте bcrypt для хеширования паролей.

## Известные ограничения

1. **Журнал событий на странице охраны**: отображается placeholder, т.к. в backend нет endpoint для получения scan_events
2. **Отчеты**: раздел отчетов показывает информацию о необходимых backend endpoints
3. **Парковка**: функционал парковки требует дополнительных endpoints

Все ограничения явно указаны в UI с описанием необходимых доработок backend.

## Troubleshooting

### Ошибка подключения к backend

- Убедитесь, что backend запущен на `http://localhost:8080`
- Проверьте настройки прокси в `vite.config.ts`
- При запуске в Docker проверьте `nginx.conf` и доступность backend по имени сервиса

### TypeScript ошибки

- Убедитесь, что установлены все зависимости: `npm install`
- Проверьте версию TypeScript: `npx tsc --version` (должна быть 5.2+)

### Не работает HID-сканер

- Убедитесь, что сканер настроен на ввод строки + Enter
- Попробуйте вручную ввести UUID и нажать Enter
- Проверьте, что поле ввода в фокусе (должно быть автоматически)

## Лицензия

[Укажите лицензию]

## Контакты

[Укажите контактную информацию]

