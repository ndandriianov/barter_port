CREATE TABLE tags
(
    name TEXT PRIMARY KEY
);

ALTER TABLE tags
    ADD CONSTRAINT tags_name_check CHECK (
        char_length(name) BETWEEN 1 AND 15
            AND name ~ '^[A-Za-zА-Яа-яЁё]+$'
        );

CREATE TABLE offer_tags
(
    offer_id  UUID NOT NULL,
    tag_name  TEXT NOT NULL,

    PRIMARY KEY (offer_id, tag_name),
    FOREIGN KEY (offer_id) REFERENCES offers (id) ON DELETE CASCADE,
    FOREIGN KEY (tag_name) REFERENCES tags (name) ON DELETE CASCADE
);

CREATE INDEX offer_tags_tag_name_idx ON offer_tags (tag_name);
