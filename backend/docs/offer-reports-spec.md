# Offer Reports Moderation

## Цель

Добавить жалобы на объявления (`offer`) с ручной модерацией администратором.

При принятии жалобы:
- объявление скрывается;
- автор объявления получает штраф `-10` очков репутации;
- штраф передается в `users` сервис через `transactional outbox` в `deals` и `inbox` в `users`.

При отклонении жалобы:
- объявление не меняется;
- жалоба помечается как отклоненная.

## Основные правила

- В каждый момент времени у одного `offer` может быть только один `offer_report` в статусе `Pending`.
- Если на `offer` уже есть `Pending`-жалоба, новая жалоба не создает новый `offer_report`.
- Вместо этого создается запись в `offer_reports_messages`, которая связывает существующий `offer_report` с автором жалобы и `message_id`.
- После того как администратор рассмотрел жалобу (`Accepted` или `Rejected`), следующая жалоба на тот же `offer` создает новый `offer_report`.
- Пока по `offer` существует хотя бы одна жалоба в статусе `Pending`, объявление блокируется для изменений и удаления.
- `message_id` хранится как внешний идентификатор сообщения без `FOREIGN KEY`, потому что сообщения живут в `chats_db`, а жалобы в `deals_db`.
- Для одного `Pending`-репорта один пользователь может приложить только одно сообщение. Это обеспечивается `PRIMARY KEY (offer_report_id, author_id)`.

## Модель данных

### deals_db

Изменения в `offers`:
- `is_hidden BOOLEAN NOT NULL DEFAULT FALSE`
- `hidden_at TIMESTAMPTZ`
- `hidden_reason TEXT`
- `modification_blocked BOOLEAN NOT NULL DEFAULT FALSE`
- `modification_blocked_at TIMESTAMPTZ`

Новые сущности:
- `offer_report_status`: `Pending`, `Accepted`, `Rejected`
- `offer_reports`
- `offer_reports_messages`
- `offer_report_penalty_outbox`

### users_db

Изменения в `users`:
- `reputation_points INTEGER NOT NULL DEFAULT 0`

Новые сущности:
- `user_reputation_events`
- `user_reputation_inbox`

## Бизнес-поток

### Создание жалобы

Запрос: `POST /offers/{offerId}/reports`

Тело:
- `messageId`

Правила:
- жаловаться на собственный `offer` нельзя;
- если `offer` скрыт, новая жалоба не создается;
- если есть `Pending`-репорт:
  - если от этого пользователя еще нет записи в `offer_reports_messages`, добавляется новая связь;
  - если запись уже есть, сервис возвращает конфликт;
- если `Pending`-репорта нет:
  - создается новый `offer_report`;
  - создается первая запись в `offer_reports_messages`;
  - в `offers` включается `modification_blocked`.

### Просмотр администратором

Администратор получает:
- список жалоб;
- карточку жалобы;
- исходный `offer`;
- список связанных `message_id`.

### Просмотр автором объявления

Запрос: `GET /offers/{offerId}/reports`

Endpoint доступен:
- автору объявления;
- администратору.

Ответ содержит:
- сам `offer`;
- список всех `offer_report` по этому объявлению;
- список связанных `message_id` для каждого репорта.

### Решение администратора

Запрос: `POST /admin/offer-reports/{reportId}/resolution`

Тело:
- `accepted`
- `comment` (optional)

Если решение `Rejected`:
- `offer_report.status = Rejected`;
- заполняются `reviewed_at`, `reviewed_by`, `resolution_comment`;
- если по `offer` больше нет `Pending`, снимается `modification_blocked`.

Если решение `Accepted`:
- `offer_report.status = Accepted`;
- заполняются `reviewed_at`, `reviewed_by`, `resolution_comment`, `applied_penalty_delta = -10`;
- `offer.is_hidden = true`;
- `offer.hidden_at = now()`;
- `offer.hidden_reason = 'offer_report_accepted'`;
- в `offer_report_penalty_outbox` пишется событие для `users`.

## Outbox / Inbox

### deals -> outbox

В одной транзакции при `Accepted`:
- обновляется `offer_report`;
- скрывается `offer`;
- записывается событие в `offer_report_penalty_outbox`.

Payload события:
- `id`
- `report_id`
- `offer_id`
- `user_id`
- `delta = -10`
- `created_at`
- `reviewed_by`

`report_id` должен быть уникален в outbox, чтобы один и тот же штраф не был опубликован дважды.

### users -> inbox

`users` сервис принимает событие в `user_reputation_inbox`, затем процессор в одной транзакции:
- пишет запись в `user_reputation_events`;
- увеличивает/уменьшает `users.reputation_points`.

Идемпотентность обеспечивается уникальным ключом на:
- `user_reputation_events (source_type, source_id)`

Для этого используется:
- `source_type = 'offer_report'`
- `source_id = report_id`

## API-правила

### Публичные offers endpoints

- `GET /offers` не должен возвращать скрытые объявления.
- `GET /offers/{offerId}` для скрытого объявления должен возвращать `404`, если запрос не от администратора или не от автора объявления.
- `GET /offers/{offerId}/reports` должен быть доступен автору объявления и администратору.
- `PATCH /offers/{offerId}` и `DELETE /offers/{offerId}` должны возвращать `403`, если:
  - `modification_blocked = true`, или
  - `is_hidden = true`.

### Админские endpoints

- `GET /admin/offer-reports?status=Pending|Accepted|Rejected`
- `GET /admin/offer-reports/{reportId}`
- `POST /admin/offer-reports/{reportId}/resolution`

### Users endpoints

- `GET /users/me` должен возвращать текущее значение `reputationPoints`.

## Ошибки

- `400`:
  - невалидный `offerId`, `reportId` или `messageId`;
  - пустое тело запроса;
  - попытка принять уже рассмотренную жалобу.
- `403`:
  - жалоба на собственный `offer`;
  - попытка обычного пользователя использовать admin endpoint;
  - попытка изменить заблокированный или скрытый `offer`.
- `404`:
  - `offer` или `report` не найден;
  - `offer` скрыт и недоступен для текущего пользователя.
- `409`:
  - этот пользователь уже приложил сообщение к текущему `Pending`-репорту;
  - жалоба уже рассмотрена другим администратором.
