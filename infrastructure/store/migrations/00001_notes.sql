-- +goose Up
CREATE TABLE example_notes (
    id text PRIMARY KEY,
    owner_id text NOT NULL,
    text text NOT NULL CHECK (octet_length(text) BETWEEN 1 AND 2048),
    created_at timestamptz NOT NULL
);
CREATE INDEX example_notes_owner ON example_notes (owner_id, id);

-- +goose Down
DROP TABLE example_notes;
