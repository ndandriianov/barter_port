DROP TABLE IF EXISTS offer_report_penalty_outbox;
DROP INDEX IF EXISTS offer_reports_one_pending_per_offer_idx;
DROP INDEX IF EXISTS offer_reports_offer_id_idx;
DROP INDEX IF EXISTS offer_reports_status_created_at_idx;
DROP TABLE IF EXISTS offer_reports_messages;
DROP TABLE IF EXISTS offer_reports;
DROP TYPE IF EXISTS offer_report_status;

ALTER TABLE offers
    DROP COLUMN IF EXISTS modification_blocked_at,
    DROP COLUMN IF EXISTS modification_blocked,
    DROP COLUMN IF EXISTS hidden_reason,
    DROP COLUMN IF EXISTS hidden_at,
    DROP COLUMN IF EXISTS is_hidden;
