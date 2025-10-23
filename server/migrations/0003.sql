-- migrate:up
DROP TABLE gmail_entries;

CREATE TABLE gmail_entries (
    user_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    thread_id TEXT NOT NULL,
    labels JSONB NOT NULL,
    snippet TEXT NOT NULL,
    history_id INTEGER NOT NULL,
    internal_date INTEGER NOT NULL, -- epoch ms
    headers JSONB NOT NULL,
    sender JSONB NOT NULL,
    receiver JSONB NOT NULL,
    received_at TEXT NOT NULL,
    reply_to TEXT NOT NULL,
    additional_receivers JSONB NOT NULL, -- eg bcc, cc
    PRIMARY KEY (user_id, message_id)
);

DROP TABLE gmail_entry_bodies;

CREATE TABLE gmail_entry_bodies (
    user_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    plain_text TEXT NOT NULL,
    html TEXT NOT NULL,
    has_attachments INTEGER NOT NULL,
    attachment_ids JSONB NOT NULL,
    PRIMARY KEY (user_id, message_id)
);

-- migrate:down
