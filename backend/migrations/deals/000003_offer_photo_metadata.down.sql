ALTER TABLE offer_photos
    DROP CONSTRAINT IF EXISTS offer_photos_offer_id_position_key;

ALTER TABLE offer_photos
    DROP CONSTRAINT IF EXISTS offer_photos_pkey;

ALTER TABLE offer_photos
    ADD CONSTRAINT offer_photos_pkey PRIMARY KEY (offer_id, url);

ALTER TABLE offer_photos
    DROP COLUMN IF EXISTS position;

ALTER TABLE offer_photos
    DROP COLUMN IF EXISTS id;
