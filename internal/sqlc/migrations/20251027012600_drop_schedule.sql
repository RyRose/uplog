-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS schedule;
DROP TABLE IF EXISTS schedule_list;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE schedule (
    date TEXT PRIMARY KEY NOT NULL CHECK (date LIKE '____-__-__'),
    workout TEXT NULL,
    notes TEXT NOT NULL,
    FOREIGN KEY (workout) REFERENCES workout (id)
);

CREATE TABLE schedule_list (
    id TEXT NOT NULL,
    day INTEGER NOT NULL CHECK (day >= 0),
    workout TEXT NOT NULL,
    PRIMARY KEY (id, day),
    FOREIGN KEY (workout) REFERENCES workout (id)
);
-- +goose StatementEnd
