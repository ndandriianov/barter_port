CREATE TABLE email_tokens
(
    token_hash TEXT PRIMARY KEY,
    user_id    UUID        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN     NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE users
(
    id             UUID PRIMARY KEY,
    email          TEXT        NOT NULL UNIQUE,
    password_hash  TEXT        NOT NULL,
    email_verified BOOLEAN     NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL
);


CREATE TABLE user_creation_outbox
(
    id         UUID PRIMARY KEY,
    event_id   UUID        NOT NULL,
    user_id    UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE user_creation_events
(
    id         UUID PRIMARY KEY,
    user_id    UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    status     TEXT        NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT user_creation_events_status_check CHECK (status IN ('NEW', 'CREATED', 'FAILED'))
);

CREATE TABLE refresh_tokens
(
    jti        TEXT PRIMARY KEY,
    user_id    UUID        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked    BOOLEAN     NOT NULL
);

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
    CONSTRAINT items_action_check CHECK (action IN ('give', 'take')),

    FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE DATABASE users_db;

\connect users_db;

CREATE TABLE users
(
    id   UUID PRIMARY KEY,
    name TEXT,
    bio  TEXT
);

CREATE TABLE user_creation_inbox
(
    id         UUID PRIMARY KEY,
    event_id   UUID        NOT NULL,
    user_id    UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE deleted_users
(
    id         UUID PRIMARY KEY,
    deleted_at TIMESTAMPTZ NOT NULL
)

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