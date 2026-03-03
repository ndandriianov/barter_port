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
    name        TEXT        NOT NULL,
    type        TEXT        NOT NULL,
    action      TEXT        NOT NULL,
    description TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    views       INTEGER     NOT NULL DEFAULT 0,

    CONSTRAINT items_type_check CHECK (type IN ('good', 'service')),
    CONSTRAINT items_action_check CHECK (action IN ('give', 'take'))
);