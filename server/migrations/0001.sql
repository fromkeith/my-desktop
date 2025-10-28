-- migrate:up
CREATE TABLE oauth_init_session (
    state TEXT NOT NULL PRIMARY KEY,
    claimed_id TEXT NOT NULL,
    code_verifier TEXT NOT NULL,
    post_auth_return TEXT,
    created_at TEXT NOT NULL,
    expires_at TEXT NOT NULL
);

CREATE TABLE oauth_token_record (
    user_id TEXT NOT NULL PRIMARY KEY,
    provider TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expiry TEXT NOT NULL,
    token_type TEXT NOT NULL,
    scope TEXT NOT NULL
);

CREATE TABLE user_accounts (
    account_id TEXT NOT NULL PRIMARY KEY,
    created_at TEXT NOT NULL
);

CREATE TABLE user_oauth_accounts (
    account_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (account_id, user_id)
);

CREATE TABLE gmail_sync_status (
    user_id TEXT NOT NULL PRIMARY KEY,
    history_id TEXT NOT NULL,
    until TEXT NOT NULL
);


CREATE TABLE gmail_entries (
    user_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    thread_id TEXT NOT NULL,
    labels JSONB NOT NULL,
    subject TEXT NOT NULL,
    snippet TEXT NOT NULL,
    history_id INTEGER NOT NULL,
    internal_date INTEGER NOT NULL, -- epoch ms
    headers JSONB NOT NULL,
    sender JSONB NOT NULL,
    receiver JSONB NOT NULL,
    received_at TEXT NOT NULL,
    reply_to JSONB NOT NULL,
    additional_receivers JSONB NOT NULL, -- eg bcc, cc
    PRIMARY KEY (user_id, message_id)
);


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
