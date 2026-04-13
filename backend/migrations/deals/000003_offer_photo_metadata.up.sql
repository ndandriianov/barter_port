ALTER TABLE offer_photos
    ADD COLUMN id UUID;

UPDATE offer_photos
SET id = gen_random_uuid()
WHERE id IS NULL;

ALTER TABLE offer_photos
    ALTER COLUMN id SET NOT NULL;

ALTER TABLE offer_photos
    ADD COLUMN position INTEGER;

WITH ranked_photos AS (
    SELECT ctid,
           row_number() OVER (PARTITION BY offer_id ORDER BY url) - 1 AS next_position
    FROM offer_photos
)
UPDATE offer_photos op
SET position = ranked_photos.next_position
FROM ranked_photos
WHERE op.ctid = ranked_photos.ctid;

ALTER TABLE offer_photos
    ALTER COLUMN position SET NOT NULL;

ALTER TABLE offer_photos
    DROP CONSTRAINT offer_photos_pkey;

ALTER TABLE offer_photos
    ADD CONSTRAINT offer_photos_pkey PRIMARY KEY (id);

ALTER TABLE offer_photos
    ADD CONSTRAINT offer_photos_offer_id_position_key UNIQUE (offer_id, position);
