# Debug test DB in GoLand

Для интеграционных тестов в режиме debug/reuse тестовый Postgres поднимается на фиксированном адресе:

- host: `127.0.0.1`
- port: `15432`
- user: `postgres`
- password: `postgres`

Если `15432` занят, можно задать другой порт через `BARTER_PORT_TEST_POSTGRES_HOST_PORT`.

## Запуск нужного теста

Из корня `backend`:

```bash
make test-integration-moderator-resolution-debug
```

Тест:

```text
TestModeratorResolutionForFailureSuccessAndReadableByParticipant
```

Контейнеры в этом режиме не удаляются после завершения теста. Чтобы очистить их вручную:

```bash
make test-integration-clean
```

## Что смотреть для этого теста

Основные базы и таблицы:

- `deals_db.public.deals`
- `deals_db.public.deal_failures`
- `deals_db.public.reputation_events_outbox`
- `users_db.public.users`
- `users_db.public.user_reputation_events`

## Подключение из GoLand

1. Откройте `View` -> `Tool Windows` -> `Database`.
2. Нажмите `+` -> `Data Source` -> `PostgreSQL`.
3. Заполните поля:
   - `Host`: `127.0.0.1`
   - `Port`: `15432`
   - `User`: `postgres`
   - `Password`: `postgres`
   - `Database`: `deals_db`
4. Нажмите `Test Connection`.
5. Сохраните datasource как `integration deals`.
6. Создайте второй datasource теми же параметрами, но с `Database = users_db`. Назовите его `integration users`.

Если выбрали другой порт через `BARTER_PORT_TEST_POSTGRES_HOST_PORT`, укажите его вместо `15432`.

## Полезные SQL-запросы

```sql
select id, status, updated_at
from deals
order by updated_at desc
limit 20;
```

```sql
select deal_id, user_id, confirmed_by_admin, punishment_points, admin_comment
from deal_failures
order by deal_id desc;
```

```sql
select id, source_type, source_id, user_id, delta, created_at, comment
from reputation_events_outbox
order by created_at desc
limit 20;
```

```sql
select id, reputation_points
from users
order by id desc
limit 20;
```

```sql
select id, user_id, source_type, source_id, delta, created_at, comment
from user_reputation_events
order by created_at desc
limit 20;
```
