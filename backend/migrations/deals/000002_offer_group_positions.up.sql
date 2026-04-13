ALTER TABLE offer_group_units
    ADD COLUMN position INTEGER NOT NULL DEFAULT 0;

WITH ordered_units AS (
    SELECT
        id,
        ROW_NUMBER() OVER (PARTITION BY offer_group_id ORDER BY ctid) - 1 AS position
    FROM offer_group_units
)
UPDATE offer_group_units AS ogu
SET position = ordered_units.position
FROM ordered_units
WHERE ogu.id = ordered_units.id;

ALTER TABLE offer_group_units
    ADD CONSTRAINT offer_group_units_offer_group_id_position_key UNIQUE (offer_group_id, position);

ALTER TABLE unit_offers
    ADD COLUMN position INTEGER NOT NULL DEFAULT 0;

WITH ordered_unit_offers AS (
    SELECT
        unit_id,
        offer_id,
        ROW_NUMBER() OVER (PARTITION BY unit_id ORDER BY ctid) - 1 AS position
    FROM unit_offers
)
UPDATE unit_offers AS uo
SET position = ordered_unit_offers.position
FROM ordered_unit_offers
WHERE uo.unit_id = ordered_unit_offers.unit_id
  AND uo.offer_id = ordered_unit_offers.offer_id;

ALTER TABLE unit_offers
    ADD CONSTRAINT unit_offers_unit_id_position_key UNIQUE (unit_id, position);
