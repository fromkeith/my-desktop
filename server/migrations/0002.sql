-- migrate:up
CREATE TABLE gmail_sync_status (
    user_id TEXT NOT NULL PRIMARY KEY,
    history_id TEXT NOT NULL,
    until TEXT NOT NULL
);

CREATE TABLE gmail_entries (
    user_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    thread_id TEXT NOT NULL,
    labels JSON NOT NULL,
    snippet TEXT NOT NULL,
    history_id INTEGER NOT NULL,
    internal_date INTEGER NOT NULL, -- epoch ms
    headers JSON NOT NULL,
    sender JSON NOT NULL,
    receiver JSON NOT NULL,
    received_at TEXT NOT NULL,
    reply_to TEXT NOT NULL,
    additional_receivers JSONB NOT NULL, -- eg bcc, cc
    PRIMARY KEY (user_id, message_id)
);

CREATE TABLE gmail_entry_bodies (
    user_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    plain_text TEXT NOT NULL,
    html TEXT NOT NULL,
    has_attachments INTEGER NOT NULL,
    attachment_ids JSON NOT NULL,
    PRIMARY KEY (user_id, message_id)
);

-- migrate:down
