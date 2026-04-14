ALTER TABLE users
    ADD COLUMN reputation_points INTEGER NOT NULL DEFAULT 0;

CREATE TABLE user_reputation_events
(
    id          UUID PRIMARY KEY,
    user_id     UUID        NOT NULL,
    source_type TEXT        NOT NULL,
    source_id   UUID        NOT NULL,
    delta       INTEGER     NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    comment     TEXT,

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX user_reputation_events_source_idx
    ON user_reputation_events (source_type, source_id);

CREATE INDEX user_reputation_events_user_created_at_idx
    ON user_reputation_events (user_id, created_at DESC);

CREATE TABLE user_reputation_inbox
(
    id          UUID PRIMARY KEY,
    source_type TEXT        NOT NULL,
    source_id   UUID        NOT NULL,
    user_id     UUID        NOT NULL,
    delta       INTEGER     NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    comment     TEXT
);

CREATE UNIQUE INDEX user_reputation_inbox_source_idx
    ON user_reputation_inbox (source_type, source_id);
