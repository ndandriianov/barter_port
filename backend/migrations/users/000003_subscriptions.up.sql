CREATE TABLE subscriptions
(
    target_user_id UUID        NOT NULL,
    subscriber_id  UUID        NOT NULL,

    FOREIGN KEY (target_user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (subscriber_id) REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (target_user_id, subscriber_id)
);
