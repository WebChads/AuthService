# AuthService

Микросервис для аутентификации, авторизации и управления токенами.

## Описание

AuthService предоставляет API для:
- Генерации и валидации JWT токенов
- Регистрации пользователей
- Верификации через SMS
- Интеграции с Kafka для событий аутентификации

## API Endpoints

### Аутентификация
- `POST /api/v1/auth/generate-token` - Генерация JWT токена (Тестовый ендпойнт для разработчиков)
- `POST /api/v1/auth/validate-token` - Валидация JWT токена

### Регистрация
- `POST /api/v1/auth/register` - Регистрация нового пользователя
- `POST /api/v1/auth/send-sms-code` - Отправка SMS с кодом подтверждения
- `POST /api/v1/auth/verify-sms-code` - Проверка SMS кода и выдача токена

### Документация
- `GET /swagger/*` - Swagger документация API

## Конфигурация

Конфигурация сервиса задается в файле `configs/appsettings.json`, который работает как для локальной разработки, так и для Docker-контейнера.

Пример конфигурации:
```json
{
    "port": "8081",
    "secret_key": "some_cool_key",
    "is_development": true,
    "database": {
        "host": "localhost:5432",
        "db_name": "auth_service_db",
        "user": "postgres",
        "password": "postgres"
    },
    "kafka": {
        "url": "localhost:9092"
    }
}
```

## Зависимости от внешних сервисов

Для работы AuthService требуются:
- PostgreSQL - основное хранилище данных
- Kafka + Zookeeper - для общения с SmsService
- [SmsService](https://github.com/WebChads/SmsService) - сервис для отправки SMS (взаимодействие через Kafka)

## Запуск

### Локальный запуск
1. Убедитесь, что все внешние сервисы запущены
2. Настройте конфигурацию под вас в файле `configs/appsettings.json`
3. Запустите сервис (зависимости должны подтянуться при билде):
```bash
go run main.go
```

### Запуск через Docker
Используйте `docker-compose` для запуска всего стека:
```bash
docker-compose up -d --build
```

После запуска документация API будет доступна по адресу `http://localhost:<PORT>/swagger/`

## Безопасность

API использует JWT для аутентификации. Токен должен передаваться в заголовке `Authorization` в формате:
```
Authorization: Bearer <token>
```