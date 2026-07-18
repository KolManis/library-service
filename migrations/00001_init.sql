-- +goose Up
CREATE TABLE books (
    id      uuid PRIMARY KEY,
    title   text NOT NULL,
    author  text NOT NULL,
    isbn    text NOT NULL
);

CREATE TABLE copies (
    id             uuid PRIMARY KEY,
    book_id        uuid NOT NULL REFERENCES books (id),
    decommissioned boolean NOT NULL DEFAULT false
);

CREATE TABLE readers (
    id        uuid PRIMARY KEY,
    full_name text NOT NULL,
    email     text NOT NULL,
    status    text NOT NULL
);

CREATE TABLE loans (
    id          uuid PRIMARY KEY,
    copy_id     uuid NOT NULL REFERENCES copies (id),
    reader_id   uuid NOT NULL REFERENCES readers (id),
    status      text NOT NULL,
    reserved_at timestamptz NOT NULL,
    issued_at   timestamptz,
    due_at      timestamptz,
    returned_at timestamptz
);

CREATE TABLE fines (
    id      uuid PRIMARY KEY,
    loan_id uuid NOT NULL REFERENCES loans (id),
    amount  numeric(10,2) NOT NULL,
    status  text NOT NULL
);

CREATE INDEX idx_loans_copy_id_status ON loans (copy_id, status);
-- +goose Down
DROP TABLE IF EXISTS fines;
DROP TABLE IF EXISTS loans;
DROP TABLE IF EXISTS copies;
DROP TABLE IF EXISTS readers;
DROP TABLE IF EXISTS books;