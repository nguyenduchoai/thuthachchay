-- +goose Up
-- +goose StatementBegin
UPDATE challenges SET status='settled' WHERE status='done';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE challenges SET status='done' WHERE status='settled';
-- +goose StatementEnd
