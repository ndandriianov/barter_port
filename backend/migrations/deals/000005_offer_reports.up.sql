ALTER TABLE offers
    ADD COLUMN is_hidden               BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN hidden_at               TIMESTAMPTZ,
    ADD COLUMN hidden_reason           TEXT,
    ADD COLUMN modification_blocked    BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN modification_blocked_at TIMESTAMPTZ;

CREATE TYPE offer_report_status AS ENUM (
    'Pending',
    'Accepted',
    'Rejected'
    );

CREATE TABLE offer_reports
(
    id                    UUID PRIMARY KEY,
    offer_id              UUID                NOT NULL,
    offer_author_id       UUID                NOT NULL,
    status                offer_report_status NOT NULL DEFAULT 'Pending',
    created_at            TIMESTAMPTZ         NOT NULL DEFAULT now(),
    reviewed_at           TIMESTAMPTZ,
    reviewed_by           UUID,
    resolution_comment    TEXT,
    applied_penalty_delta INTEGER,

    FOREIGN KEY (offer_id) REFERENCES offers (id) ON DELETE CASCADE
);

CREATE TABLE offer_reports_messages
(
    offer_report_id UUID NOT NULL,
    author_id       UUID NOT NULL,
    message         TEXT NOT NULL,

    PRIMARY KEY (offer_report_id, author_id),
    FOREIGN KEY (offer_report_id) REFERENCES offer_reports (id) ON DELETE CASCADE
);

CREATE INDEX offer_reports_status_created_at_idx
    ON offer_reports (status, created_at DESC);

CREATE INDEX offer_reports_offer_id_idx
    ON offer_reports (offer_id);

CREATE UNIQUE INDEX offer_reports_one_pending_per_offer_idx
    ON offer_reports (offer_id)
    WHERE status = 'Pending';

CREATE TABLE offer_report_penalty_outbox
(
    id          UUID PRIMARY KEY,
    report_id   UUID        NOT NULL UNIQUE,
    offer_id    UUID        NOT NULL,
    user_id     UUID        NOT NULL,
    delta       INTEGER     NOT NULL,
    reviewed_by UUID        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,

    FOREIGN KEY (report_id) REFERENCES offer_reports (id) ON DELETE CASCADE,
    FOREIGN KEY (offer_id) REFERENCES offers (id) ON DELETE CASCADE,
    CONSTRAINT offer_report_penalty_outbox_delta_check CHECK (delta < 0)
);
