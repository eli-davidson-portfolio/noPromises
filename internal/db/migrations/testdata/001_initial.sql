-- +migrate Up
CREATE TABLE test (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE test; 