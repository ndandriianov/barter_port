# Frontend Rewrite Input

## Цель документа

Этот документ фиксирует текущее функциональное покрытие фронтенда и ограничения, которые диктуются backend-спеками. Он нужен как входной материал для следующего шага: проектирования новой навигации, новых user flow и перераспределения экранов/виджетов.

Важно:

- документ не проектирует новый UX;
- документ не предлагает новую IA или новую карту переходов;
- документ описывает, что уже покрывает текущий фронтенд, где это находится, какие backend-правила это ограничивают и какие наблюдаемые проблемы есть в текущей структуре.

## Границы фронтенда

Текущий фронтенд отвечает за следующие продуктовые области:

- аутентификация и восстановление доступа;
- профиль текущего пользователя;
- публичные профили пользователей и подписки;
- каталог обычных объявлений;
- композитные объявления (`offer-groups`);
- черновики сделок;
- сделки и позиции сделок;
- личные чаты и чаты сделок;
- отзывы;
- статистика текущего пользователя;
- жалобы на объявления;
- admin-сценарии для жалоб на объявления и провалов сделок.

Вне этого документа:

- детальное проектирование нового UX;
- предложение новой структуры меню;
- приоритизация работ;
- технический план рефакторинга.

## Источники истины

### Backend specs

- `backend/docs/auth/swagger.yaml`
- `backend/docs/doc-first/users/swagger.yaml`
- `backend/docs/doc-first/users/paths/subscriptions.yaml`
- `backend/docs/doc-first/chats/swagger.yaml`
- `backend/docs/doc-first/chats/paths/chats.yaml`
- `backend/docs/doc-first/deals/swagger.yaml`
- `backend/docs/doc-first/deals/paths/offers.yaml`
- `backend/docs/doc-first/deals/paths/offer-groups.yaml`
- `backend/docs/doc-first/deals/paths/drafts.yaml`
- `backend/docs/doc-first/deals/paths/deals.yaml`
- `backend/docs/doc-first/deals/paths/joins.yaml`
- `backend/docs/doc-first/deals/paths/reviews.yaml`
- `backend/docs/doc-first/deals/paths/failures.yaml`
- `backend/docs/doc-first/deals/paths/statistics.yaml`
- `backend/docs/offer-reports-spec.md`

### Текущий frontend

- роутинг: `frontend/src/shared/config/Router.tsx`
- layout’ы: `frontend/src/widgets/layout/*`
- API-слой: `frontend/src/features/*/api/*`
- страницы: `frontend/src/pages/**`
- ключевые виджеты: `frontend/src/widgets/**`
- глобальное состояние: `frontend/src/app/store/*`, `frontend/src/features/auth/model/*`

## Технический срез текущего фронтенда

### Стек

- React 19
- TypeScript
- Vite
- Material UI
- Redux Toolkit + RTK Query
- React Router
- Zod для валидации и парсинга ответов

### Состояние и API

- Единственный явный slice состояния: `auth`, хранит `accessToken`.
- Доменные данные в основном живут в RTK Query cache.
- Для каждого домена есть отдельный `createApi`: `auth`, `users`, `offers`, `offer-groups`, `deals`, `chats`, `reviews`, `statistics`.
- `baseQueryWithReauth` автоматически пробует `POST /auth/refresh` на `401`.
- `refresh` использует cookie (`credentials: include`), а access token хранится в Redux.
- При logout и при неуспешном refresh очищается `auth` и сбрасываются RTK Query cache.

### Layout’ы

- `AuthLayout`: отдельный изолированный контейнер для `/login`, `/register`, `/verify-email`, `/reset-password`.
- `AppLayout`: общий shell приложения с верхним AppBar, бургером, навигацией и badge’ами.

### Глобальная навигация в AppLayout

Текущие верхнеуровневые ссылки:

- `/offers`
- `/offer-reports/mine`
- `/offer-groups`
- `/deals`
- `/reviews`
- `/deals/drafts`
- `/profile`
- `/chats`
- `/admin` для администратора

Дополнительные наблюдения:

- badge на `/reviews` показывает количество доступных отзывов;
- badge на `/deals/drafts` показывает количество входящих черновиков;
- `/statistics` не вынесена в верхний уровень навигации, хотя отдельная страница существует.

## Карта текущих маршрутов

| Route | Экран | Назначение |
| --- | --- | --- |
| `/login` | LoginPage | вход |
| `/register` | RegisterPage | регистрация |
| `/verify-email` | VerifyEmailPage | подтверждение email по token |
| `/reset-password` | ResetPasswordPage | сброс пароля по token |
| `/profile` | ProfilePage | свой профиль, пароль, репутация, подписки |
| `/users/:userId` | UserPage | публичный профиль другого пользователя |
| `/users/:userId/reviews` | UserReviewsPage | отзывы о пользователе как о provider |
| `/offers` | OffersListPage | каталог обычных объявлений |
| `/offers/create` | CreateOfferPage | создание объявления |
| `/offers/:offerId/edit` | EditOfferPage | редактирование своего объявления |
| `/offers/:offerId` | OfferPage | карточка объявления |
| `/offers/:offerId/reviews` | OfferReviewsPage | отзывы по конкретному offer |
| `/offer-reports/mine` | MyOfferReportsPage | жалобы на мои объявления |
| `/offer-groups` | OfferGroupsListPage | список композитных объявлений |
| `/offer-groups/create` | CreateOfferGroupPage | создание композитного объявления |
| `/offer-groups/:offerGroupId` | OfferGroupPage | карточка композитного объявления |
| `/deals` | DealsListPage | список сделок |
| `/deals/:dealId` | DealPage | карточка сделки + чат |
| `/deals/:dealId/items/:itemId` | DealItemPage | карточка позиции сделки |
| `/deals/drafts` | DraftsListPage | список черновиков |
| `/deals/drafts/:draftId` | DraftPage | карточка черновика |
| `/chats` | ChatsPage | список и окно чатов |
| `/reviews` | ReviewsPage | отзывы: доступные, мои, обо мне |
| `/statistics` | StatisticsPage | агрегированная статистика |
| `/admin` | AdminPage | вход в админские сценарии |
| `/admin/offer-reports` | AdminOfferReportsPage | очередь жалоб на объявления |
| `/admin/offer-reports/:reportId` | AdminOfferReportDetailsPage | разбор конкретной жалобы |

## Функциональный инвентарь по доменам

### 1. Auth

Текущее покрытие:

- регистрация по email и паролю;
- вход по email и паролю;
- logout;
- подтверждение email по token из query string;
- запрос письма на сброс пароля;
- установка нового пароля по token;
- смена пароля из профиля для авторизованного пользователя.

Текущие frontend точки:

- `pages/auth/*`
- `widgets/auth-form/*`
- `features/auth/api/authApi.ts`

Основные backend endpoints:

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout`
- `POST /auth/verify-email`
- `POST /auth/request-password-reset`
- `POST /auth/reset-password`
- `POST /auth/change-password`
- `POST /auth/refresh`

### 2. Current User / Profile

Текущее покрытие:

- загрузка профиля текущего пользователя;
- редактирование `name`, `bio`, `avatarUrl`, `phoneNumber`;
- загрузка аватара отдельным multipart endpoint;
- сохранение, изменение и очистка текущей точки пользователя;
- просмотр `reputationPoints`;
- просмотр истории репутационных событий в drawer;
- просмотр своих подписок и своих подписчиков;
- смена пароля;
- logout из профиля.

Особенности UI:

- профиль совмещает редактирование персональных данных, работу с локацией, пароль, подписки, репутацию и logout;
- репутационные события открываются в drawer, а не на отдельной странице.

Основные backend endpoints:

- `GET /users/me`
- `PATCH /users/me`
- `POST /users/me/avatar`
- `GET /users/reputation-events`
- `GET /users/subscriptions`
- `GET /users/subscribers`

### 3. Users / Subscriptions

Текущее покрытие:

- открытие публичного профиля другого пользователя;
- просмотр его телефона, bio и avatar;
- подписка и отписка;
- просмотр его подписок и подписчиков;
- переход к отзывам о пользователе.

Основные backend endpoints:

- `GET /users/{id}`
- `GET /users/subscriptions/{id}`
- `GET /users/subscribers/{id}`
- `POST /users/subscriptions`
- `DELETE /users/subscriptions`

Backend-правило, влияющее на будущий UX:

- direct chat можно создать только при взаимной подписке;
- после разрыва взаимной подписки новый чат создать нельзя, но существующий чат остаётся рабочим.

### 4. Offers

Текущее покрытие списка:

- вкладки: `others`, `favorites`, `subscriptions`, `mine`;
- сортировки: `ByTime`, `ByPopularity`, `ByDistance`;
- фильтрация по тегам;
- фильтр `withoutTags`;
- бесконечная догрузка страниц через cursor-based pagination;
- отображение расстояния при наличии координат.

Текущее покрытие карточки offer:

- просмотр full card;
- открытие фото;
- просмотр местоположения;
- просмотр summary отзывов по offer;
- инкремент просмотров;
- добавление/удаление из избранного;
- переход к отзывам по offer;
- переход к отзывам о provider;
- создание жалобы;
- создание draft deal через отклик;
- для автора: edit/delete;
- для автора и admin: показ moderation state;
- для автора: ссылка на черновики, связанные с offer.

Текущее покрытие создания/редактирования:

- create offer;
- edit own offer;
- теги с подсказками;
- до 10 фото;
- удаление существующих фото;
- установка или очистка координат объявления.

Текущее покрытие модерации жалоб со стороны автора:

- страница `/offer-reports/mine` показывает мои объявления и жалобы по ним;
- виден статус каждой жалобы и комментарий модератора, если он есть.

Основные backend endpoints:

- `GET /offers`
- `GET /offers/subscriptions`
- `GET /offers/favorites`
- `POST /offers`
- `GET /offers/{offerId}`
- `PATCH /offers/{offerId}`
- `DELETE /offers/{offerId}`
- `PUT /offers/{offerId}/favorite`
- `DELETE /offers/{offerId}/favorite`
- `POST /offers/{offerId}/view`
- `GET /tags`
- `GET /offers/{offerId}/reports`
- `POST /offers/{offerId}/reports`

### 5. Offer Groups

Текущее покрытие:

- список чужих и своих композитных объявлений;
- создание `offer_group` из собственных обычных offers;
- составление `units` по модели `AND` между блоками и `OR` внутри блока;
- запрет смешивать разные `action` внутри одного `unit`;
- просмотр карточки группы;
- отклик на группу с выбором одного offer из каждого `unit`;
- опциональный или обязательный `responderOffer`, в зависимости от uniform action;
- создание обычного draft deal на основе `offer_group`.

Основные backend endpoints:

- `GET /offer-groups`
- `POST /offer-groups`
- `GET /offer-groups/{offerGroupId}`
- `POST /offer-groups/{offerGroupId}/drafts`

### 6. Draft Deals

Текущее покрытие:

- создание черновика напрямую из offer detail;
- создание черновика из `offer_group`;
- список черновиков с режимами `all`, `others`, `mine`;
- фильтрация черновиков по одному из моих offers;
- карточка конкретного draft;
- confirm participation;
- cancel participation;
- delete/reject draft.

Текущая семантика в UI:

- `mine` воспринимается как исходящие черновики;
- `others` воспринимается как входящие предложения;
- `all` — объединённый список.

Основные backend endpoints:

- `POST /deals/drafts`
- `GET /deals/drafts`
- `GET /deals/drafts/{draftId}`
- `PATCH /deals/drafts/{draftId}`
- `PATCH /deals/drafts/{draftId}/cancel`
- `DELETE /deals/drafts/{draftId}`

### 7. Deals

Текущее покрытие списка:

- единый список сделок;
- фильтрация по status tab: `LookingForParticipants`, `Discussion`, `Confirmed`, `Completed`, `Cancelled`, `Failed`;
- фильтр `myOnly`;
- переход в карточку сделки.

Текущее покрытие карточки сделки:

- просмотр сделки и участников;
- редактирование названия сделки;
- добавление позиции в сделку;
- работа с join requests;
- join deal;
- leave deal;
- голосование по смене статуса;
- отображение и управление `items`;
- переход в `DealItemPage`;
- встроенный `DealFailureSection`;
- встроенный чат сделки;
- alert о доступных отзывах после `Completed`.

Текущее покрытие позиции сделки:

- просмотр item;
- просмотр фото;
- редактирование контентных полей автором позиции;
- claim/release ролей `provider` и `receiver`.

Текущее покрытие провала сделки:

- участники и admin видят голоса по провалу;
- участники могут голосовать за виновного или отзывать голос;
- admin может открыть материалы провала и принять решение;
- на admin dashboard есть очередь провалов сделок.

Основные backend endpoints:

- `GET /deals`
- `GET /deals/{dealId}`
- `PATCH /deals/{dealId}`
- `PATCH /deals/{dealId}/status`
- `GET /deals/{dealId}/status`
- `POST /deals/{dealId}/items`
- `PATCH /deals/{dealId}/items/{itemId}`
- `POST /deals/{dealId}/joins`
- `GET /deals/{dealId}/joins`
- `DELETE /deals/{dealId}/joins`
- `POST /deals/{dealId}/joins/{userId}`
- `GET /deals/failures/review`
- `GET /deals/failures/{dealId}/votes`
- `POST /deals/failures/{dealId}/votes`
- `DELETE /deals/failures/{dealId}/votes`
- `GET /deals/failures/{dealId}/materials`
- `POST /deals/failures/{dealId}/moderator-resolution`

### 8. Chats

Текущее покрытие:

- список личных чатов;
- список чатов сделок;
- ручное создание нового direct chat;
- повторное открытие уже существующего direct chat вместо дублирования;
- окно сообщений;
- периодический polling сообщений;
- отправка сообщений;
- read-only режим для некоторых чатов сделок.

Основные backend endpoints:

- `GET /chats`
- `POST /chats`
- `GET /chats/users`
- `GET /chats/deals/{dealId}`
- `GET /chats/{chatId}/messages`
- `POST /chats/{chatId}/messages`

### 9. Reviews

Текущее покрытие:

- единая страница `/reviews` с тремя вкладками:
  - `available`
  - `mine`
  - `about-me`
- сбор доступных отзывов по всем завершённым сделкам;
- фильтрация available reviews по `dealId`;
- создание отзыва на уровне конкретной позиции сделки;
- просмотр и summary отзывов по offer;
- просмотр и summary отзывов по provider;
- просмотр собственных отзывов как author;
- edit/delete собственных отзывов.

Текущая модель в UI:

- отзыв пишет получатель, оценивает provider;
- часть pending review карточек может восстанавливать `itemRef` через `deal + offerRef`, если backend вернул не item, а offer context.

Основные backend endpoints:

- `GET /offers/{offerId}/reviews`
- `GET /offers/{offerId}/reviews-summary`
- `GET /providers/{providerId}/reviews`
- `GET /providers/{providerId}/reviews-summary`
- `GET /authors/{authorId}/reviews`
- `GET /reviews/{reviewId}`
- `PATCH /reviews/{reviewId}`
- `DELETE /reviews/{reviewId}`
- `GET /deals/{dealId}/reviews`
- `GET /deals/{dealId}/reviews-pending`
- `GET /deals/{dealId}/items/{itemId}/reviews/eligibility`
- `GET /deals/{dealId}/items/{itemId}/reviews`
- `POST /deals/{dealId}/items/{itemId}/reviews`

### 10. Statistics

Текущее покрытие:

- отдельная страница персональной статистики;
- сделки: completed / failed / active;
- объявления: total / totalViews;
- отзывы: written / received / average rating;
- жалобы: filedByMe и breakdown по моим объявлениям.

Основной backend endpoint:

- `GET /me/statistics`

### 11. Admin

Текущее покрытие:

- gate по `GET /users/me` и `isAdmin`;
- admin landing page;
- список тегов и удаление тегов;
- очередь жалоб на объявления;
- карточка конкретной жалобы;
- accept/reject жалобы на объявление;
- очередь сделок, переданных на разбор провала;
- диалог принятия решения по провалу сделки.

Основные backend endpoints:

- `GET /admin/offer-reports`
- `GET /admin/offer-reports/{reportId}`
- `POST /admin/offer-reports/{reportId}/resolution`
- `DELETE /admin/tags`
- `GET /deals/failures/review`
- `GET /deals/failures/{dealId}/materials`
- `POST /deals/failures/{dealId}/moderator-resolution`

## Текущие пользовательские сценарии

Ниже перечислены не целевые, а уже существующие сценарии, которые текущий frontend реально покрывает.

### Auth scenarios

1. Регистрация:
   `/register` -> ввод email/password -> `POST /auth/register` -> success alert -> ожидание email verification.
2. Подтверждение email:
   переход по ссылке -> `/verify-email?token=...` -> `POST /auth/verify-email`.
3. Вход:
   `/login` -> `POST /auth/login` -> access token в Redux -> переход на `/`.
4. Восстановление пароля:
   dialog из login -> `POST /auth/request-password-reset` -> переход по email link -> `/reset-password?token=...` -> `POST /auth/reset-password`.
5. Смена пароля:
   `/profile` -> dialog -> `POST /auth/change-password`.

### Offer discovery and response scenarios

1. Просмотр общего каталога:
   `/offers?tab=others` -> выбор сортировки/тегов -> `GET /offers`.
2. Просмотр подписок:
   `/offers?tab=subscriptions` -> `GET /offers/subscriptions`.
3. Просмотр избранного:
   `/offers?tab=favorites` -> `GET /offers/favorites`.
4. Просмотр только своих объявлений:
   `/offers?tab=mine` -> `GET /offers?my=true`.
5. Открытие карточки объявления:
   `/offers/:offerId` -> `GET /offers/{offerId}` -> для чужого offer ещё `POST /offers/{offerId}/view`.
6. Отклик на обычное объявление:
   offer detail -> modal выбора своих offers -> `POST /deals/drafts` -> переход на `/deals/drafts/:draftId`.
7. Жалоба на объявление:
   offer detail -> modal -> `POST /offers/{offerId}/reports`.
8. Редактирование своего объявления:
   `/offers/:offerId/edit` -> `PATCH /offers/{offerId}`.
9. Удаление своего объявления:
   offer detail -> `DELETE /offers/{offerId}`.

### Offer group scenarios

1. Создание группы:
   `/offer-groups/create` -> выбор своих offers по unit -> `POST /offer-groups`.
2. Отклик на группу:
   `/offer-groups/:offerGroupId` -> выбор по одному offer из unit -> optional/required responder offer -> `POST /offer-groups/{offerGroupId}/drafts`.

### Draft scenarios

1. Просмотр списка черновиков:
   `/deals/drafts` -> `GET /deals/drafts` в разных комбинациях флагов.
2. Просмотр карточки черновика:
   `/deals/drafts/:draftId` -> `GET /deals/drafts/{draftId}`.
3. Подтверждение участия:
   `PATCH /deals/drafts/{draftId}`.
4. Отмена участия:
   `PATCH /deals/drafts/{draftId}/cancel`.
5. Удаление/отклонение черновика:
   `DELETE /deals/drafts/{draftId}`.

### Deal scenarios

1. Просмотр списка сделок:
   `/deals` -> `GET /deals`.
2. Вход в карточку сделки:
   `/deals/:dealId` -> `GET /deals/{dealId}`.
3. Join flow:
   deal detail -> `POST /deals/{dealId}/joins`.
4. Leave flow:
   deal detail -> `DELETE /deals/{dealId}/joins`.
5. Обработка join request:
   deal detail -> `POST /deals/{dealId}/joins/{userId}?accept=...`.
6. Смена статуса сделки:
   deal detail -> `PATCH /deals/{dealId}/status`.
7. Добавление позиции:
   deal detail -> add item dialog -> `POST /deals/{dealId}/items`.
8. Редактирование позиции:
   `DealItemPage` или edit dialog -> `PATCH /deals/{dealId}/items/{itemId}`.
9. Голосование за провал:
   `DealFailureSection` -> `POST /deals/failures/{dealId}/votes` или `DELETE /deals/failures/{dealId}/votes`.
10. Чат сделки:
    `DealPage` -> `GET /chats/deals/{dealId}` -> `GET/POST /chats/{chatId}/messages`.

### Review scenarios

1. Вход через completed deal:
   deal detail после `Completed` -> ссылка в `/reviews?tab=available&dealId=:id`.
2. Обзор всех доступных отзывов:
   `/reviews?tab=available` -> `GET /deals?my=true` -> для completed deals `GET /deals/{dealId}/reviews-pending`.
3. Создание отзыва:
   pending review card -> dialog -> `POST /deals/{dealId}/items/{itemId}/reviews`.
4. Просмотр своих отзывов:
   `/reviews?tab=mine` -> `GET /authors/{authorId}/reviews`.
5. Редактирование/удаление своего отзыва:
   `PATCH /reviews/{reviewId}`, `DELETE /reviews/{reviewId}`.
6. Просмотр отзывов о себе:
   `/reviews?tab=about-me` -> `GET /providers/{providerId}/reviews`.
7. Просмотр отзывов по offer:
   `/offers/:offerId/reviews`.

### Chat scenarios

1. Просмотр всех чатов:
   `/chats` -> `GET /chats`.
2. Создание direct chat:
   modal -> `GET /chats/users` -> `POST /chats`.
3. Работа с уже существующим direct chat:
   modal находит существующий чат и открывает его без нового запроса на создание.
4. Работа с чатом сделки:
   из `DealPage`, если сервер вернул `dealChat`.

### Moderation scenarios

1. Автор смотрит жалобы на свои объявления:
   `/offer-reports/mine` -> мои offers -> `GET /offers/{offerId}/reports`.
2. Админ смотрит очередь жалоб:
   `/admin/offer-reports` -> `GET /admin/offer-reports`.
3. Админ разбирает жалобу:
   `/admin/offer-reports/:reportId` -> `GET /admin/offer-reports/{reportId}` -> `POST /admin/offer-reports/{reportId}/resolution`.
4. Админ смотрит очередь провалов:
   `/admin` -> `GET /deals/failures/review`.
5. Админ разбирает провал сделки:
   dialog -> `GET /deals/failures/{dealId}/materials` + `GET /deals/failures/{dealId}/votes` -> `POST /deals/failures/{dealId}/moderator-resolution`.

## Бизнес-правила и ограничения из backend

### Аутентификация

- доступ к доменным данным везде предполагает авторизацию;
- refresh идёт через cookie, access token обновляется отдельно;
- email может быть обязательным условием успешного login.

### Видимость объявлений

- hidden offers не видны обычным пользователям;
- hidden offers видны автору в режиме `my=true`;
- hidden offers видны admin;
- `GET /offers/{offerId}` для hidden offer возвращает `404` неавтору и неадмину.

### Изменение и удаление объявлений

- редактировать и удалять может только автор;
- если `modification_blocked = true`, edit/delete запрещены;
- если `is_hidden = true`, edit/delete запрещены;
- это особенно важно для жалоб на объявления.

### Жалобы на объявления

- нельзя жаловаться на собственное объявление;
- у одного offer в каждый момент только один `Pending` report;
- если `Pending` report уже есть, новая жалоба добавляется в существующее разбирательство;
- один пользователь может приложить только одно сообщение к одному `Pending` report;
- при `Accepted` объявление скрывается и автор получает штраф `-10`;
- пока есть `Pending` report, объявление блокируется для изменений и удаления;
- автор объявления и admin могут смотреть историю жалоб по offer.

### Сделки

- не-админ видит все `LookingForParticipants`, но остальные статусы только если он participant;
- `my=true` для сделок сужает выдачу до моих участий;
- join/leave/process join зависят от статуса сделки;
- если по сделке уже есть `deal_failures` без решения админа, join/leave/process join/status change/item edits блокируются;
- переходы статусов голосуются участниками;
- `Failed` нельзя установить обычным change status endpoint;
- при переходе в `Completed` начисляется репутация асинхронно.

### Позиции сделки

- content fields позиции меняет только author item;
- claim/release provider/receiver ограничены бизнес-правилами;
- при наличии pending failure moderation редактирование item запрещено.

### Провал сделки

- голосовать можно только участнику сделки;
- голосование доступно только в `Discussion` и `Confirmed`;
- после достижения порога и создания `deal_failures` действия по сделке блокируются до решения admin;
- admin получает отдельную очередь провалов;
- admin может подтвердить провал, указать виновника и штраф, либо отклонить провал.

### Offer groups

- в `unit` можно включать только собственные offers текущего пользователя;
- внутри одного `unit` все offers должны иметь одинаковый `action`;
- при создании draft из группы нужно выбрать ровно один offer из каждого unit;
- если у всех unit одинаковый `action`, responder offer обязателен и должен иметь такой же `action`.

### Подписки и чаты

- direct chat создаётся только при взаимной подписке;
- после разрыва взаимной подписки новый чат создать нельзя;
- существующий чат после разрыва подписки сохраняется.

### Чаты сделок

- чат сделки доступен участнику сделки или admin;
- отправка сообщений запрещена в финальных статусах `Completed`, `Cancelled`, `Failed`;
- отправка сообщений также запрещена при pending failure moderation.

### Отзывы

- offer reviews доступны любому авторизованному пользователю;
- provider reviews доступны любому авторизованному пользователю;
- available reviews завязаны на completed deals;
- оценка относится к фактическому provider;
- summary endpoints возвращают нули, если отзывов нет.

## Связь backend -> frontend

Ниже перечислена связь на уровне доменов. Это не полная копия swagger, а карта того, где endpoint участвует во frontend.

| Endpoint group | Где используется во frontend | Какой сценарий обслуживает |
| --- | --- | --- |
| `/auth/*` | auth pages, profile | login/register/verify/reset/change password/logout |
| `/users/me`, `/users/me/avatar`, `/users/reputation-events` | profile | управление своим профилем |
| `/users/subscriptions*`, `/users/subscribers*` | profile, user page | social graph и условия для direct chat |
| `/offers`, `/offers/subscriptions`, `/offers/favorites`, `/tags` | offers list, create/edit offer | discovery и фильтрация |
| `/offers/{offerId}`, `/offers/{offerId}/view`, `/offers/{offerId}/favorite` | offer page | просмотр карточки, view count, favorites |
| `/offers/{offerId}/reports`, `/admin/offer-reports*` | offer page, my offer reports, admin pages | complaint flow |
| `/offer-groups*` | offer group pages | composite offer flow |
| `/deals/drafts*` | respond modals, drafts pages | draft creation and confirmation |
| `/deals`, `/deals/{dealId}`, `/deals/{dealId}/status` | deals list, deal page | основная lifecycle модель сделки |
| `/deals/{dealId}/items*` | deal page, deal item page | item management |
| `/deals/{dealId}/joins*` | deal page | join request flow |
| `/deals/failures*` | deal page, admin page | failure voting and moderation |
| `/chats*` | chats page, deal page | direct chats и deal chats |
| `/reviews*`, `/offers/*/reviews*`, `/providers/*/reviews*`, `/authors/*/reviews` | review pages, offer page, user page, deal page | review lifecycle |
| `/me/statistics` | statistics page | personal metrics |

## Наблюдаемые проблемные места текущего фронта

Этот раздел фиксирует проблемы текущей структуры. Он не содержит нового решения.

### 1. Глобальная навигация не совпадает с естественными продуктовыми сценариями

- в верхнем уровне вместе смешаны объекты разной природы: каталог, жалобы на мои объявления, композиты, сделки, отзывы, черновики, профиль, чаты;
- `/statistics` есть как отдельный продуктовый экран, но не встроена в основной navigation flow;
- админские сценарии partly isolated, partly embedded в общий deal flow.

### 2. Один бизнес-процесс часто разрезан на несколько разнородных экранов

- flow “откликнуться на объявление -> договориться -> участвовать в сделке -> оставить отзыв” распределён между `OfferPage`, modal, `DraftPage`, `DealPage`, `DealItemPage`, `/reviews`;
- flow “жалоба на объявление” распределён между `OfferPage`, `/offer-reports/mine`, `/admin/offer-reports`, admin detail;
- flow “провал сделки” распределён между `DealPage`, `DealFailureSection`, admin dashboard dialog.

### 3. Сильная зависимость от локальных modal-потоков

- отклик на обычный offer живёт в modal;
- отклик на offer group живёт в modal;
- создание direct chat живёт в modal;
- complaint creation живёт в modal;
- review creation живёт в dialog;
- часть ключевых действий не имеет собственного экранного контекста и теряется внутри detail pages.

### 4. Несколько доменов пересекаются в одних и тех же экранах

- `DealPage` одновременно является карточкой сделки, entry point для join flow, status flow, item flow, failure flow, chat flow и review reminders;
- `ProfilePage` одновременно редактирование профиля, смена пароля, просмотр социальных связей, просмотр репутации и logout;
- `AdminPage` одновременно overview, tags management и очередь провалов сделок.

### 5. Состояния доступа и бизнес-ограничений часто отображаются только условными кнопками и alert’ами

- hidden / modification blocked для offer;
- read-only state чата сделки;
- доступность голосования по провалу;
- доступность отзывов после completed deal;
- условия responder offer для offer group.

Это означает, что часть реального пользовательского сценария сейчас выражена не через отдельные шаги, а через набор условных элементов на карточках.

### 6. Разные разделы используют разные UI-подходы

- большая часть приложения построена на MUI;
- chats реализованы на inline styles и визуально/структурно живут отдельно от остального интерфейса;
- это усиливает ощущение, что чаты не встроены в общую продуктовую модель.

### 7. Query-param navigation используется как substitute для экранной структуры

- `/offers?tab=...`
- `/offer-groups?tab=...`
- `/deals?status=...`
- `/deals/drafts?tab=...`
- `/reviews?tab=...`

Это делает текущий frontend функционально покрытым, но слабо собранным в понятную информационную архитектуру.

## Открытые вопросы и риски до проектирования новых user flow

### 1. Несовпадение complaint flow между текстовой спекой и текущим фронтом

`backend/docs/offer-reports-spec.md` описывает создание жалобы через `messageId`, привязанный к сообщению в chat. Текущий frontend отправляет обычный текст `message` через `POST /offers/{offerId}/reports`.

Для следующего шага нужно явно зафиксировать, какая модель является актуальной:

- жалоба через свободный текст;
- жалоба через ссылку на сообщение из чата;
- или гибрид.

### 2. Не до конца выражен продуктовый статус draft -> deal

Текущий код различает:

- draft как входящее предложение;
- draft как исходящий сценарий;
- deal как уже подтверждённую сущность.

Но в UI это разложено между отдельными страницами и модалками без единой “лестницы состояний”.

### 3. Неочевидно, что считать primary entry point для reviews

Сейчас попасть к отзывам можно:

- из главного nav;
- из alert на completed deal;
- из offer page;
- из user page.

Документ фиксирует все входы, но следующий шаг должен решить, какой из них основной, а какие вспомогательные.

### 4. Failure moderation частично встроена в deal UI, частично в admin UI

Нужно отдельно решить на следующем шаге, где проходит граница между:

- participant-side failure flow;
- admin-side moderation flow.

### 5. Statistics существует как отдельный экран, но не является частью явного продукта в navigation

Это влияет на будущую продуктовую карту, но не должно решаться внутри этого документа.

## Что должно сохраниться при переделке

При redesign/rewrite нельзя потерять:

- refresh-based auth с access token + cookie refresh;
- профиль, avatar upload, phone, bio, location;
- subscriptions и subscribers;
- полный каталог offers с режимами `others/favorites/subscriptions/mine`;
- create/edit/delete offer;
- favorites;
- distance/popularity/time sorting;
- offer reports для пользователя и для admin;
- hidden/modification-blocked moderation states;
- offer groups и draft creation из них;
- draft deals и все их действия;
- deals, joins, status voting, items;
- failure voting и admin failure resolution;
- direct chats;
- deal chats с read-only ограничениями;
- review lifecycle;
- personal statistics;
- admin entry gate по `isAdmin`.

Также нельзя потерять backend-driven ограничения:

- visibility rules;
- ownership rules;
- role-based access;
- pending failure locks;
- moderation locks;
- mutual subscription prerequisite for direct chat.

## Материалы для следующего шага

Этот README должен использоваться как исходник для:

1. построения новой продуктовой карты экранов;
2. определения primary и secondary сценариев в каждом домене;
3. проектирования новой навигационной модели;
4. решения, какие действия остаются в modal/dialog, а какие становятся самостоятельными экранами;
5. определения новых user flow без потери backend coverage;
6. составления плана перераспределения текущих widget/page responsibility.

Следующий шаг должен опираться на три вещи из этого документа:

- функциональный инвентарь;
- текущие сценарии;
- backend-ограничения и открытые противоречия.
