ALTER TABLE offer_report_penalty_outbox
    DROP CONSTRAINT IF EXISTS offer_report_penalty_outbox_report_id_fkey,
    DROP CONSTRAINT IF EXISTS offer_report_penalty_outbox_offer_id_fkey,
    DROP CONSTRAINT IF EXISTS offer_report_penalty_outbox_delta_check,
    ALTER COLUMN report_id DROP NOT NULL,
    ALTER COLUMN reviewed_by DROP NOT NULL;

ALTER TABLE offer_report_penalty_outbox
    RENAME TO reputation_events_outbox;

ALTER TABLE reputation_events_outbox
    RENAME COLUMN offer_id TO source_id;

ALTER TABLE reputation_events_outbox
    ADD COLUMN source_type TEXT,
    ADD COLUMN comment TEXT;

UPDATE reputation_events_outbox
SET source_type = 'deals.offer_report.penalty'
WHERE source_type IS NULL;

ALTER TABLE reputation_events_outbox
    ALTER COLUMN source_type SET NOT NULL;

CREATE INDEX reputation_events_outbox_created_at_id_idx
    ON reputation_events_outbox (created_at, id);
