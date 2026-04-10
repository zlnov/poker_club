# Poker Club Backend - MVP

Backend на Go с использованием фреймворка Gin для системы отслеживания домашних покерных игр.

## Архитектура

Проект следует принципам Clean Architecture:

- **Domain Layer** (`/domain`) - сущности, агрегаты, value objects, доменные сервисы
- **Infrastructure Layer** (`/infrastructure`) - репозитории, подключение к БД
- **Application Layer** (`/application`) - use cases, DTOs
- **Presentation Layer** (`/presentation`) - HTTP handlers, middleware

## Требования

- Go 1.21+
- PostgreSQL 14+
- Git

## Быстрый старт

### 1. Установка зависимостей

```bash
cd backend
go mod download
```

### 2. Настройка базы данных

Создайте базу данных PostgreSQL:

```sql
CREATE DATABASE poker_club;
CREATE USER poker_user WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE poker_club TO poker_user;
```

### 3. Настройка переменных окружения

Скопируйте `.env.example` в `.env` и при необходимости измените:

```bash
cp .env.example .env
```

Параметры `.env`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=poker_user
DB_PASSWORD=password
DB_NAME=poker_club
DB_SSLMODE=disable

SERVER_PORT=8080
```

### 4. Применение миграций

Миграции находятся в папке `migrations/`. Выполните:

```bash
psql -U poker_user -d poker_club -f migrations/001_initial_schema.sql
```

Или используйте любой клиент PostgreSQL для выполнения SQL-файла.

### 5. Запуск сервера

```bash
go run main.go
```

Сервер запустится на `http://localhost:8080`

### 6. Проверка работы

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:
```json
{"status":"ok"}
```

## API Endpoints

### Clubs

- `POST /api/v1/clubs` - Создать клуб
- `GET /api/v1/clubs/:club_id/members` - Получить участников клуба
- `POST /api/v1/clubs/:club_id/members/approve` - Утвердить заявку участника
- `POST /api/v1/clubs/:club_id/members/reject` - Отклонить заявку участника

### Games

- `POST /api/v1/games` - Создать игру
- `GET /api/v1/games/:game_id/participants` - Получить участников игры
- `POST /api/v1/games/:game_id/buyin` - Зарегистрировать buy-in/rebuy
- `POST /api/v1/games/:game_id/participants/:player_id/chips` - Установить количество фишек
- `POST /api/v1/games/:game_id/finish` - Завершить игру
- `GET /api/v1/players/:player_id/stats` - Получить статистику игрока

## Примеры использования

### 1. Создание клуба

```bash
curl -X POST http://localhost:8080/api/v1/clubs \
  -H "Content-Type: application/json" \
  -d '{"name":"My Poker Club"}'
```

### 2. Создание игры

```bash
curl -X POST http://localhost:8080/api/v1/games \
  -H "Content-Type: application/json" \
  -d '{
    "club_id": 1,
    "banker_id": 1,
    "type": "cash_open",
    "money_model": "real",
    "buy_in_amount": 1000,
    "rebuy_allowed": true,
    "rebuy_amount": 1000,
    "max_rebuys_per_player": 3,
    "start_time": "2024-01-01T19:00:00Z",
    "min_players": 2,
    "max_players": 10,
    "ranking_primary": "chips",
    "ranking_secondary": "place"
  }'
```

### 3. Buy-in игрока

```bash
curl -X POST http://localhost:8080/api/v1/games/1/buyin \
  -H "Content-Type: application/json" \
  -d '{"player_id": 1}'
```

### 4. Установка фишек и завершение игры

```bash
# Установить фишки для игрока 1
curl -X POST http://localhost:8080/api/v1/games/1/participants/1/chips \
  -H "Content-Type: application/json" \
  -d '{"chips": 1500}'

# Завершить игру
curl -X POST http://localhost:8080/api/v1/games/1/finish \
  -H "Content-Type: application/json" \
  -d '{}'
```

## Структура проекта

```
backend/
├── application/
│   ├── dtos/           # Data Transfer Objects
│   │   ├── requests.go
│   │   └── responses.go
│   └── usecases/       # Use cases (application services)
│       ├── club_usecase.go
│       └── game_usecase.go
├── domain/             # Domain layer
│   ├── models.go       # Entity models
│   ├── repositories.go # Repository interfaces
│   ├── services.go     # Domain services
│   └── value_objects.go # Value objects & enums
├── infrastructure/
│   └── persistence/    # Data persistence
│       ├── database.go
│       ├── models.go   # GORM models
│       └── repositories.go # Repository implementations
├── presentation/
│   ├── handlers/       # HTTP handlers
│   │   ├── club_handler.go
│   │   └── game_handler.go
│   └── middleware/     # Middleware
│       └── auth.go
├── migrations/         # Database migrations
│   └── 001_initial_schema.sql
├── main.go            # Application entry point
├── go.mod
├── .env.example
└── README-START-MVP.md
```

## Важные замечания

### MVP ограничения

1. **Аутентификация**: В MVP используется заглушка - `user_id` всегда равен 1. В продакшн нужно реализовать JWT или другую аутентификацию.
2. **Транзакции**: Все операции используют транзакции через GORM.
3. **Event Log**: Все изменения состояния записываются в таблицу `events`.
4. **Валидация**: Используется `go-playground/validator` для валидации входных данных.
5. **Блокировки**: Для предотвращения гонок используется `SELECT ... FOR UPDATE` в `LockForUpdate`.

### Бизнес-правила

- **SUM(chips_end) == SUM(invested)** - проверяется при завершении игры
- **Только активные участники** могут участвовать в играх
- **Только admin** может утверждать/отклонять заявки
- **Rebuy** разрешен только если `rebuy_allowed = true`
- **Max rebuys** проверяется при каждом rebuy

## Разработка

### Запуск с hot-reload (рекомендуется)

Установите air:

```bash
go install github.com/cosmtrek/air@latest
```

Запуск:

```bash
air
```

### Тестирование

```bash
go test ./...
```

### Линтинг

```bash
go vet ./...
gofmt -l .
```

## Проблемы и решения

### Ошибка "could not import github.com/gin-gonic/gin"

Убедитесь, что зависимости загружены:

```bash
go mod download
go mod tidy
```

### Ошибка подключения к БД

Проверьте:
1. PostgreSQL запущен: `pg_isready`
2. База данных создана
3. Параметры в `.env` корректны

### Миграции уже применены

Если вы видите ошибку о существующих таблицах, миграции уже были применены (GORM AutoMigrate).
## Дальнейшее развитие Backend

- [ ] Реализовать JWT аутентификацию
- [ ] Добавить Webhook для интеграции с Telegram ботом
- [ ] Реализовать кэширование частых запросов (Redis)
- [ ] Добавить более детальную аналитику
- [ ] Написать unit и integration тесты
- [ ] Добавить Docker конфигурацию
- [ ] Настроить CI/CD



## Frontend часть проекта

Frontend реализован на React с использованием TypeScript, Vite и Tailwind CSS. Он предоставляет пользовательский интерфейс для взаимодействия с покерным клубом.

### Технологический стек

- **React 18** - библиотека для построения пользовательских интерфейсов
- **TypeScript** - типизированный superset JavaScript
- **Vite** - быстрый сборщик и dev-сервер
- **Tailwind CSS** - утилитарный CSS-фреймворк для стилизации
- **React Query** - управление состоянием серверных данных
- **Zustand** - легковесное управление состоянием клиента
- **React Router** - навигация и роутинг
- **Axios** - HTTP клиент для коммуникации с бэкендом
- **Lucide React** - набор иконок
- **Recharts** - библиотека для построения графиков
- **Date-fns** - библиотека для работы с датами
- **JWT-decode** - декодирование JWT токенов
- **Class Variance Authority** и **clsx** - управление условными классами
- **Radix UI Primitives** - доступные UI компоненты

### Архитектура

Frontend следует модульной структуре с разделением ответственности:

```
frontend/
├── src/
│   ├── app/                    # Компоненты layout и общие компоненты приложения
│   ├── assets/                 # Статические ресурсы (изображения, шрифты)
│   ├── entities/               # Типы и интерфейсы сущностей (TypeScript interfaces)
│   ├── features/               # Feature-slices с логикой по доменам (auth, clubs, games, players)
│   │   ├── auth/               # Аутентификация и авторизация
│   │   ├── clubs/              # Управление клубами
│   │   ├── games/              # Управление играми
│   │   └── players/            # Управление игроками
│   ├── pages/                  # Страницы приложения
│   ├── shared/                 # Переиспользуемые компоненты и утилиты
│   │   ├── components/         # Переиспользуемые UI компоненты
│   │   ├── lib/                # Утилиты и хелперы (API, auth, utils)
│   │   └── ui/                 # Базовые UI компоненты (button, input, card и т.д.)
│   └── widgets/                # Визуальные виджеты (графики, диаграммы)
├── public/                     # Статические файлы, обслуживаемые как есть
├── index.html                  # Точка входа HTML
├── vite.config.ts              # Конфигурация Vite
├── tailwind.config.js          # Конфигурация Tailwind CSS
├── tsconfig.json               # Конфигурация TypeScript
├── postcss.config.js           # Конфигурация PostCSS
├── package.json                # Зависимости и скрипты
└── .env.example                # Пример файла переменных окружения
```

### Состояния и управление данными

- **React Query** используется для кэширования серверных данных, автоматической пере валидации фоновых обновлений и управления состоянием запросов
- **Zustand** с persist middleware используется для хранения данных аутентификации (пользователь, токены) в localStorage
- Каждый feature-slice содержит свои собственные хуки для работы с API и управления локальным состоянием

### API коммуникация

Frontend взаимодействует с бэкендом через REST API:
- Базовый URL настраивается через переменную окружения `VITE_API_URL`
- Все запросы автоматически добавляют Authorization header с Bearer токеном
- Реализовано автоматическое обновление access токена через refresh токен при получении 401 ошибки
- Используется axios с кастомными интерцепторами для обработки ошибок и токенов

### Ключевые особенности UI/UX

- Адаптивный дизайн с боковой панелью для навигации
- Темная тема поддержка через класс `dark` на корневом элементе
- Защищенные маршруты через компонент `ProtectedRoute`
- Уведомления и обратная связь пользователю
- Формы с валидацией и обработкой ошибок
- Интернационализация дат через date-fns с русской локалью
- Доступные компоненты на основе Radix UI Primitives

### Быстрый старт Frontend

#### 1. Установка зависимостей

```bash
cd frontend
npm install
```

#### 2. Настройка переменных окружения

Скопируйте `.env.example` в `.env` и при необходимости измените:

```bash
cp .env.example .env
```

Параметры `.env`:
```env
VITE_API_URL=http://localhost:8080
```

#### 3. Запуск разработческого сервера

```bash
npm run dev
```

Frontend будет доступен по адресу `http://localhost:5173`

#### 4. Сборка для продакшена

```bash
npm run build
```

Собранные файлы будут размещены в директории `dist/`

#### 5. Предпросмотр продакшен сборки

```bash
npm run preview
```

### Взаимодействие с бэкендом

Для полноценной работы frontend требует запущенного бэкенда:
1. Запустите бэкенд согласно инструкциям в разделе "Быстрый старт" выше
2. Убедитесь, что бэкенд доступен по адресу, указанному в `VITE_API_URL` (по умолчанию `http://localhost:8080`)
3. Запустите frontend разработческий сервер
4. Войдите в систему используя учетные данные (в текущей MVP реализации используется заглушка аутентификации)

### Текущие ограничения MVP

1. **Аутентификация**: В текущей реализации используется упрощенная аутентификация с заглушкой. В продакшн версии планируется реализовать полноценную JWT аутентификацию с refresh токенами.
2. **Отсутствующие эндпоинты**: Некоторые страницы временно скрыты из-за отсутствия соответствующих эндпоинтов в бэкенде (например, страница списка игроков).
3. **Обработка ошибок**: Базовая обработка ошибок реализована, но требует дальнейшего улучшения для продакшена.

