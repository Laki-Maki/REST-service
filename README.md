# REST-Service: Subscription Aggregation

[![Go](https://img.shields.io/badge/Go-1.21-blue?logo=go&logoColor=white)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue?logo=docker&logoColor=white)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13-blue?logo=postgresql&logoColor=white)](https://www.postgresql.org/)


https://github.com/user-attachments/assets/150aa1cd-d355-4ba0-9518-8da8d55348c0


---

## Описание

**REST-Service** — это микросервис для агрегации данных о подписках пользователей на онлайн-сервисы.  
Сервис позволяет:

- Создавать, читать, обновлять и удалять записи о подписках (CRUDL)
- Считать суммарную стоимость подписок за указанный период с фильтрацией по пользователю и названию сервиса
- Хранить данные в PostgreSQL с миграциями
- Логировать операции
- Поддерживать конфигурацию через `.env` или `.yaml`
- Предоставлять документацию API через Swagger
- Запускаться через Docker Compose

---


Каждая запись о подписке содержит:

- `service_name` — название сервиса
- `price` — стоимость подписки в рублях (целое число)
- `user_id` — UUID пользователя
- `start_date` — дата начала подписки (месяц и год, формат `MM-YYYY`)
- `end_date` — дата окончания подписки (опционально)


Пример тела запроса на создание подписки:

```json
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025"
}
```


Структура проекта
```
subscription-service/
├── .env
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── test_PS.txt
├── cmd/
│   └── server/
│       └── main.go
├── docs/
│   └── swagger.yaml
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── db/
│   │   ├── postgres.go
│   │   └── migrations/
│   │       └── 001_create_subscriptions.sql
│   ├── handler/
│   │   └── subscription.go
│   ├── model/
│   │   └── subscription.go
│   └── service/
│       └── subscription.go

```
Установка и запуск
Клонирование репозитория
```
git clone https://github.com/Laki-Maki/REST-service.git
cd REST-service
```

Настройка конфигурации

Создайте файл .env или .yaml с настройками подключения к базе данных:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=subscriptions
```
Запуск через Docker (рекомендуется)

Проект полностью собран и запускается через Docker Compose:
```
docker compose up -d --build
```

Проверить логи:
```
docker compose logs -f
```
Запуск без Docker
```
go run ./cmd/subscription
```
API - CRUDL Endpoints
| Метод  | URL                 | Описание                |
| ------ | ------------------- | ----------------------- |
| GET    | /subscriptions      | Получить все подписки   |
| GET    | /subscriptions/{id} | Получить подписку по ID |
| POST   | /subscriptions      | Создать подписку        |
| PUT    | /subscriptions/{id} | Обновить подписку       |
| DELETE | /subscriptions/{id} | Удалить подписку        |


Расчет суммарной стоимости
Метод	URL	Параметры	Описание
```
GET	/subscriptions/aggregate	from, to, user_id (опц.), service_name (опц.)	Суммарная стоимость подписок за период
```
Пример запроса:
```
GET /subscriptions/aggregate?from=07-2025&to=11-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba
```
Примеры запросов

#Для удобства проверки функционала все запросы собраны в файле test_PS.txt


Пример создания подписки через curl:
```
curl -X POST http://localhost:8080/subscriptions \
-H "Content-Type: application/json" \
-d '{
  "service_name": "Netflix",
  "price": 500,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "09-2025"
}'
```
Логи

Сервис логирует все CRUDL-операции и ошибки для удобства отладки и аудита.

Swagger

Документация доступна после запуска сервиса по адресу:

```
http://localhost:8080/swagger/index.html
```
