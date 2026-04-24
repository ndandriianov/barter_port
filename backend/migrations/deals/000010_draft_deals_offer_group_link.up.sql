ALTER TABLE draft_deals
    ADD COLUMN offer_group_id UUID;

ALTER TABLE draft_deals
    ADD CONSTRAINT draft_deals_offer_group_id_fkey
        FOREIGN KEY (offer_group_id) REFERENCES offer_groups (id) ON DELETE SET NULL;

CREATE INDEX draft_deals_offer_group_id_idx
    ON draft_deals (offer_group_id)
    WHERE offer_group_id IS NOT NULL;
