CREATE TABLE items
(
    id          UUID PRIMARY KEY,
    author_id   UUID        NOT NULL,
    name        TEXT        NOT NULL,
    type        TEXT        NOT NULL,
    action      TEXT        NOT NULL,
    description TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    views       INTEGER     NOT NULL DEFAULT 0,

    CONSTRAINT items_type_check CHECK (type IN ('good', 'service')),
    CONSTRAINT items_action_check CHECK (action IN ('give', 'take'))
);

CREATE DATABASE auth_db;
CREATE DATABASE users_db;

-- CREATE TABLE offers
-- (
--     id          UUID PRIMARY KEY,
--     item_id     UUID,
--     deal_id     UUID        NOT NULL,
--     name        TEXT        NOT NULL,
--     description TEXT,
--     created_at  TIMESTAMPTZ NOT NULL,
--     receiver_id UUID,
--     sender_id   UUID,
--
--     FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE SET NULL,
--     FOREIGN KEY (receiver_id) REFERENCES users (id) ON DELETE SET NULL,
--     FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE SET NULL
-- );
--
-- CREATE TABLE deals
-- (
--     id     UUID PRIMARY KEY,
--     status TEXT NOT NULL,
--
--     CONSTRAINT deals_status_check CHECK (
--         status IN (
--                    'SEARCHING_FOR_PARTICIPANTS',
--                    'DISCUSSION',
--                    'APPROVED',
--                    'COMPLETED',
--                    'CANCELLED',
--                    'FAILED'
--             )
--         )
-- );
