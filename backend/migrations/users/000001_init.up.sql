CREATE TABLE users
(
    id         UUID PRIMARY KEY,
    name       TEXT,
    bio        TEXT,
    avatar_url TEXT
);

CREATE TABLE user_creation_inbox
(
    id         UUID PRIMARY KEY,
    user_id    UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE user_creation_result_outbox
(
    id         UUID PRIMARY KEY,
    user_id    UUID        NOT NULL,
    status     TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT user_creation_result_outbox_status_check CHECK (status IN ('Success', 'Failed'))
);

CREATE TABLE deleted_users
(
    id         UUID PRIMARY KEY,
    deleted_at TIMESTAMPTZ NOT NULL
)
