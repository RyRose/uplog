-- +goose Up
-- +goose StatementBegin
ALTER TABLE lift ADD COLUMN
lift_group TEXT
REFERENCES lift_group (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE lift
DROP COLUMN lift_group;
-- +goose StatementEnd

-- sqlfluff:dialect:sqlite
