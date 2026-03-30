CREATE TABLE items
(
    id          UUID PRIMARY KEY,
    author_id   UUID        NOT NULL,
    name        TEXT        NOT NULL,
    type        TEXT        NOT NULL,
    action      TEXT        NOT NULL,
    description TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    views       INTEGER     NOT NULL DEFAULT 0,

    CONSTRAINT items_type_check CHECK (type IN ('good', 'service')),
    CONSTRAINT items_action_check CHECK (action IN ('give', 'take'))
);

CREATE TABLE draft_deals
(
    id          UUID PRIMARY KEY,
    author_id   UUID        NOT NULL,
    name        TEXT,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE draft_deal_items
(
    draft_deal_id UUID    NOT NULL,
    item_id       UUID    NOT NULL,
    quantity      INTEGER NOT NULL,

    FOREIGN KEY (draft_deal_id) REFERENCES draft_deals (id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE,
    PRIMARY KEY (draft_deal_id, item_id)
);