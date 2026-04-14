# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файл из `migrations/0001_create_tasks.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `POST /api/v1/tasks/{id}` - генерация экземпляров периодической задачи в диапазоне дат
- `DELETE /api/v1/tasks/{id}`

## Периодические задачи

Задача поддерживает поле `recurrence` с типами:

- `daily` — каждый `interval_days` день
- `monthly_day` — в определенный день месяца `monthly_day` (1..30)
- `specific_dates` — только на перечисленные даты `specific_dates`
- `even_days` — только по четным дням месяца
- `odd_days` — только по нечетным дням месяца

Создание/редактирование шаблона периодической задачи происходит через стандартные `POST /api/v1/tasks` и `PUT /api/v1/tasks/{id}`.

Для генерации задач используется `POST /api/v1/tasks/{id}`:

```json
{
  "from_date": "2026-04-01",
  "to_date": "2026-04-15"
}
```

Сервис создаст задачи-экземпляры на подходящие дни и не создаст дубликаты для одной и той же даты.

## Принятые допущения

- Периодическая задача хранится как шаблон (`recurrence != null`), а рабочие задачи создаются отдельными сущностями.
- Экземпляры получают статус `new` и ссылаются на шаблон через `source_task_id`.
- Генерация выполняется явно API-методом (а не фоновым cron), чтобы поведение было предсказуемым и управляемым.
- Основной список `GET /api/v1/tasks` возвращает только исходные задачи/шаблоны (без сгенерированных экземпляров), чтобы не смешивать шаблоны и операционные задачи.
