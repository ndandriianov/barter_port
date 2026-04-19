DROP INDEX IF EXISTS reputation_events_outbox_created_at_id_idx;

DELETE FROM reputation_events_outbox
WHERE source_type IS DISTINCT FROM 'deals.offer_report.penalty'
   OR report_id IS NULL
   OR reviewed_by IS NULL;

ALTER TABLE reputation_events_outbox
    DROP COLUMN IF EXISTS comment,
    DROP COLUMN IF EXISTS source_type;

ALTER TABLE reputation_events_outbox
    RENAME COLUMN source_id TO offer_id;

ALTER TABLE reputation_events_outbox
    ALTER COLUMN report_id SET NOT NULL,
    ALTER COLUMN reviewed_by SET NOT NULL,
    ADD CONSTRAINT offer_report_penalty_outbox_delta_check CHECK (delta < 0),
    ADD CONSTRAINT offer_report_penalty_outbox_report_id_fkey
        FOREIGN KEY (report_id) REFERENCES offer_reports (id) ON DELETE CASCADE,
    ADD CONSTRAINT offer_report_penalty_outbox_offer_id_fkey
        FOREIGN KEY (offer_id) REFERENCES offers (id) ON DELETE CASCADE;

ALTER TABLE reputation_events_outbox
    RENAME TO offer_report_penalty_outbox;
