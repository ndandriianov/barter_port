DROP INDEX IF EXISTS draft_deals_offer_group_id_idx;

ALTER TABLE draft_deals
    DROP CONSTRAINT IF EXISTS draft_deals_offer_group_id_fkey;

ALTER TABLE draft_deals
    DROP COLUMN IF EXISTS offer_group_id;
