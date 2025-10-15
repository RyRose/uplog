-- +goose Up
-- +goose StatementBegin
INSERT OR REPLACE INTO template_variable VALUES ('PUSH', 'TODO');
INSERT OR REPLACE INTO template_variable VALUES ('PULL', 'TODO');
INSERT OR REPLACE INTO template_variable VALUES ('CORE', 'TODO');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Irreversible... good luck!

-- +goose StatementEnd

-- sqlfluff:dialect:sqlite
