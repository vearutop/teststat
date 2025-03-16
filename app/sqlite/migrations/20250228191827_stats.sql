-- +goose Up
-- +goose StatementBegin
CREATE TABLE packages
(
    `hash`    INTEGER      NOT NULL PRIMARY KEY,
    `package` VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE tests
(
    `hash`         INTEGER NOT NULL PRIMARY KEY,
    `package_hash` INTEGER NOT NULL default 0,
    `test`         VARCHAR(255)     DEFAULT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE totals
(
    `test_hash`       INTEGER NOT NULL PRIMARY KEY,
    `first_rev`       integer not null default 0,
    `last_rev`        integer not null default 0,
    `first_ums`       integer not null default 0, -- unixepoch in milliseconds
    `last_ums`        integer not null default 0, -- unixepoch in milliseconds
    `last_failed_ums` integer not null default 0, -- unixepoch in milliseconds
    `failed`          integer not null default 0,
    `passed`          integer not null default 0,
    `unfinished`      integer not null default 0,
    `skipped`         integer not null default 0,
    `output_lines`    integer not null default 0,
    `data_races`      integer not null default 0,
    `pauses`          integer not null default 0,
    `runs`            integer not null default 0,
    `cached`          integer not null default 0,
    `elapsed`         real    not null default 0
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE outputs
(
    `test_hash` INTEGER NOT NULL PRIMARY KEY, -- sample output is one per test
    `run_hash`  integer not null default 0,
    `output`    TEXT
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE revisions
(
    `hash`     INTEGER      NOT NULL PRIMARY KEY,
    `revision` VARCHAR(255) NOT NULL -- could be JSON
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE runs
(
    `hash`         INTEGER NOT NULL PRIMARY KEY, -- test_hash + revision_hash + started
    `test_hash`    INTEGER NOT NULL,
    `rev_hash`     INTEGER not null,
    `started_ums`  integer not null default 0,   -- unixepoch in milliseconds
    `result`       char(1)          default '-',
    `output_lines` integer not null default 0,
    `pauses`       integer not null default 0,
    `cached`       integer not null default 0,
    `elapsed`      real    not null default 0
);
-- +goose StatementEnd