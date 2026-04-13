CREATE TABLE items_photos
(
    id       UUID PRIMARY KEY,
    item_id   UUID    NOT NULL,
    url      TEXT    NOT NULL,
    position INTEGER NOT NULL,

    FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE,
    CONSTRAINT items_photos_item_id_position_key UNIQUE (item_id, position)
);
