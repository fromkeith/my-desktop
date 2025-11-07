-- migrate:up

DROP TABLE IF EXISTS gmail_entries;
DROP TABLE IF EXISTS gmail_entry_bodies;

CREATE TABLE people_sync_status (
    user_id TEXT NOT NULL PRIMARY KEY,
    next_sync_token TEXT NOT NULL
);


-- migrate:down
