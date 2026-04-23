ALTER TABLE users
    DROP COLUMN IF EXISTS current_latitude,
    DROP COLUMN IF EXISTS current_longitude;
