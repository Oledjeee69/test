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

Также сервис при старте применяет идемпотентную схему БД (через `CREATE ... IF NOT EXISTS`), поэтому может запускаться без docker-init скриптов (например, на Railway).

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

Публичное демо (Railway): `https://test-production-f090.up.railway.app/swagger/`

В OpenAPI по умолчанию выбран сервер Railway; для локального запуска в Swagger переключите **Servers** на `http://localhost:8080`.

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
- В перечисленных типах периодичности из задания нет отдельного «каждую неделю»; такие сценарии можно задать через `specific_dates` или расширить API позже.
- Даты в генерации и в `scheduled_for` считаются в **UTC**; для `even_days` / `odd_days` чётность берётся от **числа месяца** (1–31), а не от «каждого второго дня подряд».
- Для `monthly_day` в месяцах короче 28–31 дня соответствующего числа может не быть — такие месяцы пропускаются при генерации.

## Деплой на Railway (минимально)

- Создай проект из GitHub-репозитория.
- Добавь PostgreSQL plugin/service в Railway.
- Для backend сервиса задай переменные:
  - `DATABASE_DSN=<railway-postgres-url>`
  - `PORT` выставляется Railway автоматически.
  - `HTTP_ADDR` можно не задавать (сервис возьмет `:$PORT`).

## Где проверить схему БД

Схема описана в репозитории и применяется при старте приложения.

- **Файл миграции (источник правды для compose):** `migrations/0001_create_tasks.up.sql` — таблица `tasks`, индексы, поля `recurrence`, `source_task_id`, `scheduled_for`.
- **Применение без docker-init:** при старте сервиса выполняется тот же DDL идемпотентно — см. `ensureSchema` в `cmd/api/main.go`.

**Локально (Docker Compose), psql внутри контейнера Postgres:**

```bash
docker compose exec postgres psql -U postgres -d taskservice -c "\d tasks"
```

**Railway:** открой сервис PostgreSQL → **Query** / **Connect** и выполни:

```sql
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'tasks'
ORDER BY ordinal_position;
```

или `\d tasks`, если есть консоль с psql.

## Примеры запросов (curl)

Подставь свой хост вместо `BASE` (`http://localhost:8080` или публичный URL Railway).

```bash
BASE=http://localhost:8080

curl -s "$BASE/api/v1/tasks"

curl -s -X POST "$BASE/api/v1/tasks" \
  -H "Content-Type: application/json" \
  -d '{"title":"Шаблон","description":"тест","status":"new","recurrence":{"type":"daily","interval_days":2}}'

curl -s "$BASE/api/v1/tasks/1"

curl -s -X POST "$BASE/api/v1/tasks/1" \
  -H "Content-Type: application/json" \
  -d '{"from_date":"2026-04-01","to_date":"2026-04-10"}'

curl -s -X PUT "$BASE/api/v1/tasks/1" \
  -H "Content-Type: application/json" \
  -d '{"title":"Обновлено","description":"","status":"in_progress"}'

curl -s -X DELETE "$BASE/api/v1/tasks/1"
```

## Postman

Импортируй коллекцию: `postman/TaskService.postman_collection.json`. В переменных коллекции задай `baseUrl` (локально или публичный URL деплоя) и при необходимости `taskId` после создания шаблона.
