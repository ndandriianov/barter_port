CREATE TABLE reputation_events_outbox
(
    id          UUID PRIMARY KEY,
    source_type TEXT        NOT NULL,
    source_id   UUID        NOT NULL,
    user_id     UUID        NOT NULL,
    delta       INTEGER     NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    comment     TEXT
);

DROP TABLE IF EXISTS offer_report_penalty_outbox;