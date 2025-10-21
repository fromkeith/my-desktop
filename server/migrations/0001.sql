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

-- migrate:down
