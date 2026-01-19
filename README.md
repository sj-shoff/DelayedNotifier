# Delayed Notifier

Сервис для отложенной отправки уведомлений через очереди сообщений.

## Возможности

- Создание отложенных уведомлений
- Автоматическая отправка в указанное время
- Повторные попытки при ошибках
- Поддержка каналов: Email и Telegram
- Веб-интерфейс для управления
- Хранение в PostgreSQL + кэширование в Redis
- Очереди сообщений через RabbitMQ

## Технологии

- **Backend**: Go
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Broker**: RabbitMQ
- **Frontend**: HTML, CSS, JavaScript
- **Containerization**: Docker

## Быстрый старт

### Требования
- Docker и Docker Compose

### Запуск

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd delayed-notifier
```

2. Запустите приложение:
```bash
make docker-up
```

3. Откройте веб-интерфейс:
```
http://localhost:8031
```

## API Endpoints

### Создание уведомления
```http
POST /api/v1/notify
Content-Type: application/json

{
  "user_id": "user@example.com",
  "channel": "email",
  "message": "Текст уведомления",
  "send_at": "2024-01-01T12:00:00Z"
}
```

### Получение статуса
```http
GET /api/v1/notify/{id}
```

### Отмена уведомления
```http
DELETE /api/v1/notify/{id}
```

### Список уведомлений
```http
GET /api/v1/notifications
```

## Настройка окружения

Создайте файл `.env`

## Использование

### Через веб-интерфейс
1. Откройте `http://localhost:8031`
2. Заполните форму:
   - Получатель (email или Telegram ID)
   - Канал отправки
   - Текст сообщения
   - Дата и время отправки
3. Нажмите "Создать уведомление"

### Через API
```bash
# Создание уведомления
curl -X POST http://localhost:8031/api/v1/notify \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user@example.com",
    "channel": "email", 
    "message": "Test notification",
    "send_at": "2024-12-01T10:00:00Z"
  }'

# Проверка статуса
curl http://localhost:8031/api/v1/notify/{id}

# Отмена уведомления  
curl -X DELETE http://localhost:8031/api/v1/notify/{id}
```

## База данных

Миграции выполняются автоматически при запуске через Goose.

Структура таблицы `notifications`:
- `id` - UUID уведомления
- `user_id` - ID получателя
- `channel` - Канал отправки (email/telegram)
- `message` - Текст уведомления
- `send_at` - Время отправки
- `status` - Статус (pending/sent/cancelled/failed)
- `retries` - Количество попыток отправки
- `created_at`, `updated_at` - Временные метки

## Архитектура

1. **HTTP Handler** - Принимает запросы на создание уведомлений
2. **Message Broker** - Отложенная доставка через RabbitMQ с delayed exchange
3. **Consumer** - Обработка сообщений из очереди
4. **Notifier** - Отправка через выбранный канал
5. **Repository** - Работа с данными (PostgreSQL + Redis cache)

## Команды разработки

```bash
# Запуск без Docker
make run

# Сборка
make build

# Миграции БД
make migrate-up
make migrate-down

# Тестирование API
make curl-test

# Остановка контейнеров
make docker-down
```

## Мониторинг

- RabbitMQ Management: `http://localhost:15672` (guest/guest)
- Логи приложения: вывод в консоль с structured logging

## Статусы уведомлений

- ⏳ `pending` - Ожидает отправки
- ✅ `sent` - Успешно отправлено
- ❌ `cancelled` - Отменено пользователем
- ⚠️ `failed` - Ошибка отправки после всех попыток

## Разработка

Для локальной разработки убедитесь, что установлены:
- Go 1.24+
- PostgreSQL 15+
- Redis 7+
- RabbitMQ 3.12+