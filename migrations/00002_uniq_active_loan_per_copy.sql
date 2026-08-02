-- +goose Up
CREATE UNIQUE INDEX uniq_active_loan_per_copy
    ON loans (copy_id)
    WHERE status IN ('reserved', 'issued');

-- +goose Down
DROP INDEX IF EXISTS uniq_active_loan_per_copy;