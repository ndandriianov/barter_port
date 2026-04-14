DROP INDEX IF EXISTS user_reputation_inbox_source_idx;
DROP TABLE IF EXISTS user_reputation_inbox;
DROP INDEX IF EXISTS user_reputation_events_user_created_at_idx;
DROP INDEX IF EXISTS user_reputation_events_source_idx;
DROP TABLE IF EXISTS user_reputation_events;

ALTER TABLE users
    DROP COLUMN IF EXISTS reputation_points;
