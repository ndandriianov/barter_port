CREATE TABLE offers
(
    id          UUID PRIMARY KEY,
    author_id   UUID        NOT NULL,
    name        TEXT        NOT NULL,
    type        TEXT        NOT NULL,
    action      TEXT        NOT NULL,
    description TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ,
    views       INTEGER     NOT NULL DEFAULT 0,

    CONSTRAINT offers_type_check CHECK (type IN ('good', 'service')),
    CONSTRAINT offers_action_check CHECK (action IN ('give', 'take')
        )
);

CREATE TABLE draft_deals
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    author_id   UUID        NOT NULL,
    name        TEXT,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ
);

CREATE TABLE draft_deal_offers
(
    draft_deal_id UUID    NOT NULL,
    offer_id      UUID    NOT NULL,
    quantity      INTEGER NOT NULL,
    confirmed     BOOLEAN NOT NULL DEFAULT FALSE,

    FOREIGN KEY (draft_deal_id) REFERENCES draft_deals (id) ON DELETE CASCADE,
    FOREIGN KEY (offer_id) REFERENCES offers (id) ON DELETE CASCADE,
    PRIMARY KEY (draft_deal_id, offer_id)
);

CREATE TYPE deal_status AS ENUM (
    'LookingForParticipants',
    'Discussion',
    'Confirmed',
    'Completed',
    'Cancelled',
    'Failed'
    );

CREATE TABLE deals
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    name        TEXT,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ,
    status      deal_status NOT NULL DEFAULT 'LookingForParticipants'
);

CREATE TABLE items
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id     UUID    NOT NULL,
    offer_id    UUID,
    author_id   UUID    NOT NULL,
    provider_id UUID,
    receiver_id UUID,
    name        TEXT    NOT NULL,
    description TEXT    NOT NULL,
    type        TEXT    NOT NULL,
    updated_at  TIMESTAMPTZ,
    quantity    INTEGER NOT NULL DEFAULT 1,

    CONSTRAINT offers_type_check CHECK (type IN ('good', 'service')),
    FOREIGN KEY (deal_id) REFERENCES deals (id) ON DELETE CASCADE,
    FOREIGN KEY (offer_id) REFERENCES offers (id) ON DELETE SET NULL
);

CREATE TABLE participants
(
    deal_id                UUID NOT NULL REFERENCES deals (id) ON DELETE CASCADE,
    user_id                UUID NOT NULL,
    requested_status       deal_status,
    failure_blame_vote_for UUID,

    FOREIGN KEY (deal_id, failure_blame_vote_for)
        REFERENCES participants (deal_id, user_id)
        ON DELETE SET NULL (failure_blame_vote_for),

    PRIMARY KEY (deal_id, user_id)
);

CREATE TABLE join_requests
(
    user_id UUID NOT NULL,
    deal_id UUID NOT NULL REFERENCES deals (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, deal_id)
);

CREATE TABLE join_requests_votes
(
    user_id  UUID NOT NULL,
    deal_id  UUID NOT NULL REFERENCES deals (id) ON DELETE CASCADE,
    voter_id UUID NOT NULL,

    FOREIGN KEY (user_id, deal_id) REFERENCES join_requests (user_id, deal_id) ON DELETE CASCADE,
    FOREIGN KEY (deal_id, voter_id) REFERENCES participants (deal_id, user_id) ON DELETE CASCADE,

    PRIMARY KEY (user_id, deal_id, voter_id)
);

CREATE TABLE deal_failures
(
    deal_id            UUID NOT NULL REFERENCES deals (id) ON DELETE CASCADE,
    user_id            UUID,    -- если пользователи проголосовали за разных участников, то null
    confirmed_by_admin BOOLEAN, -- null - не подтверждено, true - подтверждено, false - отклонено
    admin_comment      TEXT,
    punishment_points  INTEGER,

    FOREIGN KEY (deal_id, user_id)
        REFERENCES participants (deal_id, user_id)
        ON DELETE SET NULL (user_id),

    PRIMARY KEY (deal_id)
);

CREATE TABLE deal_reviews
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    deal_id     UUID        NOT NULL,
    item_id     UUID,
    offer_id    UUID,
    author_id   UUID        NOT NULL,
    provider_id UUID        NOT NULL,
    rating      INTEGER     NOT NULL,
    comment     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ,

    CONSTRAINT deal_reviews_rating_check CHECK (rating BETWEEN 1 AND 5),
    CONSTRAINT deal_reviews_target_check CHECK (offer_id IS NOT NULL OR item_id IS NOT NULL),
    FOREIGN KEY (deal_id, author_id) REFERENCES participants (deal_id, user_id) ON DELETE CASCADE,
    FOREIGN KEY (deal_id) REFERENCES deals (id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE SET NULL,
    FOREIGN KEY (offer_id) REFERENCES offers (id) ON DELETE CASCADE,
    FOREIGN KEY (deal_id, provider_id) REFERENCES participants (deal_id, user_id)
);

CREATE UNIQUE INDEX deal_reviews_unique_offer_only_context_idx
    ON deal_reviews (deal_id, offer_id, author_id)
    WHERE offer_id IS NOT NULL AND item_id IS NULL;

CREATE UNIQUE INDEX deal_reviews_unique_item_only_context_idx
    ON deal_reviews (deal_id, item_id, author_id)
    WHERE offer_id IS NULL AND item_id IS NOT NULL;

CREATE UNIQUE INDEX deal_reviews_unique_offer_item_context_idx
    ON deal_reviews (deal_id, offer_id, item_id, author_id)
    WHERE offer_id IS NOT NULL AND item_id IS NOT NULL;

CREATE INDEX deal_reviews_offer_created_at_idx
    ON deal_reviews (offer_id, created_at DESC)
    WHERE offer_id IS NOT NULL;

CREATE INDEX deal_reviews_item_created_at_idx
    ON deal_reviews (item_id, created_at DESC)
    WHERE item_id IS NOT NULL;
