-- migrate:up
CREATE TABLE OauthInitSession (
    state varchar NOT NULL PRIMARY KEY,
    claimedId varchar NOT NULL,
    codeVerifier varchar NOT NULL,
    postAuthReturn varchar,
    createdAt timestamp without time zone NOT NULL,
    expiresAt timestamp without time zone NOT NULL
);

CREATE TABLE OauthTokenRecord (
    userId varchar NOT NULL PRIMARY KEY,
    provider varchar NOT NULL,
    accessToken varchar NOT NULL,
    refreshToken varchar NOT NULL,
    expiry timestamp without time zone NOT NULL,
    tokenType varchar NOT NULL,
    scope varchar NOT NULL,
    updatedAt timestamp without time zone NOT NULL,
    version BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE UserAccounts (
    accountId varchar NOT NULL PRIMARY KEY,
    createdAt timestamp without time zone NOT NULL
);

CREATE TABLE UserOauthAccounts (
    accountId varchar NOT NULL,
    userId varchar NOT NULL UNIQUE,
    PRIMARY KEY (accountId, userId)
);

CREATE TABLE GmailSyncStatus (
    userId varchar NOT NULL PRIMARY KEY,
    historyId varchar NOT NULL,
    until timestamp without time zone NOT NULL,
    lastSyncTime timestamp without time zone NOT NULL
);

CREATE TABLE PeopleSyncStatus (
    userId varchar NOT NULL PRIMARY KEY,
    nextSyncToken varchar NOT NULL,
    lastSyncTime timestamp without time zone NOT NULL
);




-- migrate:down

DROP TABLE OauthInitSession;
DROP TABLE OauthTokenRecord;
DROP TABLE UserAccounts;
DROP TABLE UserOauthAccounts;
DROP TABLE GmailSyncStatus;
DROP TABLE PeopleSyncStatus;
