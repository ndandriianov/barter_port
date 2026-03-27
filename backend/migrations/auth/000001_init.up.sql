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
    user_id    UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE user_creation_events
(
    user_id    UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    status     TEXT        NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT user_creation_events_status_check CHECK (status IN ('New', 'Success', 'Failed'))
);

CREATE TABLE user_creation_result_inbox
(
    id         UUID PRIMARY KEY,
    user_id    UUID        NOT NULL,
    status     TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT user_creation_result_inbox_status_check CHECK (status IN ('Success', 'Failed'))
);

CREATE TABLE refresh_tokens
(
    jti        TEXT PRIMARY KEY,
    user_id    UUID        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked    BOOLEAN     NOT NULL
);