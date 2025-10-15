-- +goose Up
-- +goose StatementBegin
ALTER TABLE lift_groups RENAME TO lift_group;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE lift_group RENAME TO lift_groups;
-- +goose StatementEnd

-- sqlfluff:dialect:sqlite
