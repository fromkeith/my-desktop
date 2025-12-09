-- migrate:up

ALTER TABLE GmailSyncStatus ADD COLUMN pubSubExp BIGINT NOT NULL DEFAULT 0;
ALTER TABLE GmailSyncStatus ADD COLUMN pubSubCreated TIMESTAMP NOT NULL DEFAULT '0001-01-01T00:00:00Z';

CREATE TABLE UserEmails (
    emailAddress varchar NOT NULL PRIMARY KEY,
    accountId varchar NOT NULL,
    userId varchar NOT NULL,
    primaryAddress bool NOT NULL DEFAULT false
);
-- quick list of my email addresses
CREATE INDEX idx_user_id ON UserEmails (accountId, userId);

-- migrate:down

ALTER TABLE GmailSyncStatus DROP COLUMN pubSubExp;
ALTER TABLE GmailSyncStatus DROP COLUMN pubSubCreated;
DROP TABLE UserEmails;
