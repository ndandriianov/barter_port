# Веб-приложение для бартерного обмена товарами и услугами
### Краткое название *"Barter Port"*

---

## Предназначение программы

Программа «Barter Port» предназначена для частных лиц, желающих обмениваться товарами и услугами без денежных расчетов. Пользователи смогут:
Создавать объявления о товарах и услугах, которые они предлагают или ищут.
- Искать подходящие варианты обмена по категориям, местоположению или по пользователям.
- Вести переговоры через встроенные чаты.
- Заключать сделки и фиксировать успешный обмен.
- Получать баллы и повышать уровень в рамках геймификации.

---

## Функциональные требования

- Регистрация и хранение информации о пользователях
- Возможность создания и хранения объявлений о товарах и услугах
- Управление сделками: участие пользователей, статус сделки, подтверждение условий
- Ведение чатов между участниками сделки
- Учёт жалоб и апелляций
- Система начисления очков за активность пользователей
- Черный список для пользователей

---

## Предварительная схема базы данных

```mermaid
erDiagram
    ITEM {
        uuid id PK
        string description
    }

    GOOD {
        uuid item_id PK, FK
        string good_specific_info
    }

    SERVICE {
        uuid item_id PK, FK
        string service_specific_info
    }
    


    DEAL_STATUSES {
        int id PK
        string name
    }

    USERS {
        uuid id PK
        string first_name
        string last_name
        string email
        string avatar_url
        int reputation
        timestamp reputation_calculated_at
        int level
        timestamp level_calculated_at
        timestamp registered_at
    }

    CHATS {
        uuid id PK
        string name
        timestamp created_at
        uuid dealId FK
    }

    DEAL {
        uuid id PK
        int status_id FK
    }

    OFFER {
        uuid id PK
        uuid deal_id FK
        uuid sender_id FK
        uuid receiver_id FK
        uuid item_id FK
    }
    
    DRAFTS {
        uuid id PK
        uuid user_id FK
    }
    
    DRAFT_ITEM_TYPES {
        int id PK
        string name
    }
    
    DRAFT_ITEMS {
        uuid id PK
        uuid draft_id FK
        uuid item_id FK
        int draft_item_type_id FK
    }

    CHAT_PARTICIPANTS {
        uuid chat_id PK, FK
        uuid user_id PK, FK
        timestamp joined_at
        timestamp left_at
    }
    
    BLACKLIST {
        uuid initiator_id PK, FK
        uuid target_id PK, FK
        bool restrict_messages
        bool restrict_apply_for_initiators_deals
        bool hide_targets_posts
        bool hide_initiators_posts
    }

    USERS ||--o{ CHAT_PARTICIPANTS : "участвует в чатах"
    CHATS ||--o{ CHAT_PARTICIPANTS : "включает участников"


    DEAL o|--o| CHATS : "чат может быть связан со сделкой"
    DEAL }o--|| DEAL_STATUSES : "имеет статус"
    DEAL ||--|{ OFFER : "содержит предлагаемую позицию"

    OFFER ||--|| ITEM : "ссылка на товар/услугу"
    ITEM ||--o| GOOD : "является товаром"
    ITEM ||--o| SERVICE : "является услугой"

    USERS ||--o{ OFFER : "sender"
    USERS ||--o{ OFFER : "receiver"
    
    
    USERS ||--o{ DRAFTS : "имеет черновики"
    DRAFTS ||--o{ DRAFT_ITEMS : "содержит элементы"
    DRAFT_ITEMS o{--|| ITEM : "ссылка на товар/услугу"
    DRAFT_ITEMS o{--|| DRAFT_ITEM_TYPES : "имеет тип (отдать/получить)"
    
    USERS ||--o{ HISTORY_ITEMS : "у пользователя есть история действий, на основании которой вычисляются репутация и уровень"
    BLACKLIST o{--|| USERS : "initiator_id"
    BLACKLIST o{--|| USERS : "target_id"
```

```mermaid
erDiagram
    USERS {
        uuid id PK
    }

    HISTORY_ITEMS {
        uuid id PK
        uuid user_id FK
        string item_type
        timestamp received_at
    }

    VIOLATIONS {
        uuid history_item_id PK, FK
        string violation_type
        int reputation_decrease
        int level_decrease
        string description
        timestamp cancelled_at
    }

    DEAL_FAILURES {
        uuid violation_id PK, FK
        uuid corresponding_deal_id FK
    }

    USER_REPORTS {
        uuid violation_id PK, FK
        uuid reporter_id FK
        bool is_chat_visible_for_admin
    }

    ADMIN_PUNISHMENTS {
        uuid violation_id PK, FK
    }

    DEAL_SUCCEEDS {
        uuid history_item_id PK, FK
        uuid corresponding_deal_id FK
        int reputation_increase
        int level_increase
    }

    ACHIEVEMENTS {
        uuid history_item_id PK, FK
        int type_id FK
    }

    ACHIEVEMENT_TYPES {
        int id PK
        string name
        int level_increase
    }

    APPEALS {
        uuid id PK
        uuid violation_id FK
        timestamp created_at
        string description
        int status_id FK
    }

    APPEAL_STATUSES {
        int id PK
        string name
    }

    USERS ||--o{ HISTORY_ITEMS : owns

    HISTORY_ITEMS ||--|| VIOLATIONS : is_violation
    HISTORY_ITEMS ||--|| DEAL_SUCCEEDS : is_deal_succeed
    HISTORY_ITEMS ||--|| ACHIEVEMENTS : is_achievement

    VIOLATIONS ||--|| DEAL_FAILURES : deal_failure
    VIOLATIONS ||--|| USER_REPORTS : user_report
    VIOLATIONS ||--|| ADMIN_PUNISHMENTS : admin_punishment

    ACHIEVEMENTS }o--|| ACHIEVEMENT_TYPES : type
    APPEALS }o--|| APPEAL_STATUSES : status
    APPEALS }o--|| VIOLATIONS : refers_to
```

DEAL_FAILURES здесь - событие, означающее, что другие участники сделки сочли его виновным в срыве.
Найти информацию о всех сорвавшихся сделках (в том числе случаи, когда обвинений не было предъявлено) можно
в таблице DEALS, отфильтровав по статусу 

---

## (Текстовые) ограничения на данные

- Каждый ITEM является либо товаром, либо услугой, но не одновременно
- Набор допустимых действий над сделкой (DEAL) функционально определяется её статусом
(DEAL_STATUSES)
- Один и тот же пользователь не может быть и sender, и receiver для одного и того же item
- Если чат связан со сделкой, то его набор участников определяется набором участников сделки
- Поля reputation и level в Users выполняют функцию кэша, реальные репутация и уровень
определяются через наследников HISTORY_ITEMS для конкретного пользователя
- VIOLATIONS / DEAL_SUCCEEDS / ACHIEVEMENTS
- Сделка должна содержать не менее двух записей в OFFER
  
  Каждое из этих событий не может существовать без HISTORY_ITEM и является специализацией
  одного события истории
- DEAL_FAILURES(corresponding_deal_id) должен соответствовать id проваленной сделки
- USERS.email уникален
- BLACKLIST: initiator_id != target_id

---

## Функциональные и многозначные зависимости

### Функциональные зависимости

#### USERS

- USERS.id ->
  first_name,
  last_name,
  email,
  avatar_url,
  reputation,
  reputation_calculated_at,
  level,
  level_calculated_at,
  registered_at

- USERS.email -> id, first_name, last_name, avatar_url, reputation, reputation_calculated_at,
  level, level_calculated_at, registered_at

#### ITEM / GOOD / SERVICE

- ITEM.id -> description

- GOOD.item_id -> good_specific_info

- SERVICE.item_id -> service_specific_info

- ITEM.id -> (принадлежность к GOOD или SERVICE)

#### DEAL_STATUSES

- DEAL_STATUSES.id -> name

#### DEAL

- DEAL.id -> status_id

- DEAL.id -> DEAL_STATUSES.name

#### OFFER

- OFFER.id ->
  sender_id,
  receiver_id,
  item_id

- (OFFER.id, sender_id) != receiver_id

#### CHATS

- CHATS.id ->
  name,
  created_at,
  dealId

#### CHAT_PARTICIPANTS

- (chat_id, user_id) ->
  joined_at,
  left_at

#### DRAFTS

- DRAFTS.id -> user_id

#### DRAFT_ITEM_TYPES

- DRAFT_ITEM_TYPES.id -> name

#### DRAFT_ITEMS

- DRAFT_ITEMS.id ->
  draft_id,
  item_id,
  draft_item_type_id

#### BLACKLIST

- (initiator_id, target_id) ->
  restrict_messages,
  restrict_apply_for_initiators_deals,
  hide_targets_posts,
  hide_initiators_posts

#### HISTORY_ITEMS

- HISTORY_ITEMS.id ->
  user_id,
  item_type,
  received_at

#### VIOLATIONS

- VIOLATIONS.history_item_id ->
  violation_type,
  reputation_decrease,
  level_decrease,
  description,
  cancelled_at

#### DEAL_FAILURES

- DEAL_FAILURES.violation_id -> corresponding_deal_id

- DEAL_FAILURES.corresponding_deal_id -> DEAL.status_id = (при котором name = failed)

#### USER_REPORTS

- USER_REPORTS.violation_id ->
  reporter_id,
  is_chat_visible_for_admin

#### DEAL_SUCCEEDS

- DEAL_SUCCEEDS.history_item_id ->
  corresponding_deal_id,
  reputation_increase,
  level_increase

#### ACHIEVEMENTS

- ACHIEVEMENTS.history_item_id -> type_id

#### ACHIEVEMENT_TYPES

- ACHIEVEMENT_TYPES.id ->
  name,
  level_increase

#### APPEALS

- APPEALS.id ->
  violation_id,
  created_at,
  description,
  status_id

#### APPEAL_STATUSES

- APPEALS.id ->
  violation_id,
  created_at,
  description,
  status_id

###  Многозначные зависимости отсутствуют

---

## Нормализация предварительной схемы

#### 1НФ

Все атрибуты атомарные -> в схеме все поля атомарные -> 1НФ выполняется

#### 2НФ

В каждой таблице выполняется: все неключевые поля зависят от всего ключа, а не части
-> 2НФ выполняется

#### 3НФ

Во всех таблицах отсутствуют транзитивные зависимости -> 3НФ выполняется

#### BCNF

Во всех ФЗ левая часть является суперключом -> BCNF выполняется
