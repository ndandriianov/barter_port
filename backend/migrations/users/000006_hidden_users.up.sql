CREATE TABLE hidden_users
(
    owner_user_id  UUID        NOT NULL,
    hidden_user_id UUID        NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_hidden_users_owner
        FOREIGN KEY (owner_user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_hidden_users_hidden
        FOREIGN KEY (hidden_user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT pk_hidden_users
        PRIMARY KEY (owner_user_id, hidden_user_id)
);

CREATE INDEX idx_hidden_users_owner_created_at
    ON hidden_users (owner_user_id, created_at DESC, hidden_user_id);
