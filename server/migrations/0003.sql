-- migrate:up

ALTER TABLE people_sync_status
    ADD COLUMN last_sync_time TEXT NOT NULL DEFAULT '';

ALTER TABLE gmail_sync_status
    ADD COLUMN last_sync_time TEXT NOT NULL DEFAULT '';

-- migrate:down
