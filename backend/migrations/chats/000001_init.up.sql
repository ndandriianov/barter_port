CREATE TABLE chats
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    deal_id     UUID, -- если null, то это личный чат между двумя участниками
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ
    -- FOREIGN KEY (deal_id) REFERENCES deals (id) ON DELETE CASCADE
    -- по grpc отправляется запрос на удаление чата при удалении сделки, каскадное удаление не работает из-за того, что разные сервисы используют разные базы данных, и в них нет прямой связи между таблицами
);

CREATE TABLE chat_participants
(
    chat_id UUID NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE messages
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    chat_id     UUID NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    sender_id   UUID NOT NULL,
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ
);

