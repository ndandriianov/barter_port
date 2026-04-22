CREATE TABLE favorite_offers
(
    user_id    UUID        NOT NULL,
    offer_id   UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, offer_id),
    FOREIGN KEY (offer_id) REFERENCES offers (id) ON DELETE CASCADE
);

CREATE INDEX favorite_offers_user_created_at_idx
    ON favorite_offers (user_id, created_at DESC, offer_id DESC);

CREATE INDEX favorite_offers_offer_id_idx
    ON favorite_offers (offer_id);
