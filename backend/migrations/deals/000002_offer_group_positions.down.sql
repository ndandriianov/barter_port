ALTER TABLE unit_offers
    DROP CONSTRAINT IF EXISTS unit_offers_unit_id_position_key;

ALTER TABLE unit_offers
    DROP COLUMN IF EXISTS position;

ALTER TABLE offer_group_units
    DROP CONSTRAINT IF EXISTS offer_group_units_offer_group_id_position_key;

ALTER TABLE offer_group_units
    DROP COLUMN IF EXISTS position;
