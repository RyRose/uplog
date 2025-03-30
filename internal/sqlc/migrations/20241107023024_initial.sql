-- +goose Up
-- +goose StatementBegin

-- Enable foreign keys since they're used in the schema.
PRAGMA foreign_keys = ON;

CREATE TABLE progress (
    -- An auto-generated identifier corresponding to the ROWID.
    -- See the link below for more details:
    -- https://www.sqlite.org/lang_createtable.html#rowid
    id INTEGER PRIMARY KEY NOT NULL,
    -- The lift that progress was made on.
    lift TEXT NOT NULL,
    -- The date that progress was made on. It is formatted as YYYY-MM-DD.
    date TEXT NOT NULL CHECK (date LIKE '____-__-__'),
    -- The amount of weight that was lifted according to the side_weight.
    -- For example, if the weight was 10 lbs and the side_weight is x2, then
    -- the true weight is 20 lbs.
    weight REAL NOT NULL CHECK (weight >= 0),
    -- The number of sets that were performed.
    sets INTEGER NOT NULL CHECK (sets >= 0),
    -- The number of reps per set that were performed.
    reps INTEGER NOT NULL CHECK (reps >= 0),
    -- The formula to apply to `weight` to get the full weight.
    side_weight TEXT NULL,
    FOREIGN KEY (lift) REFERENCES lift (id),
    FOREIGN KEY (side_weight) REFERENCES side_weight (id)
);

CREATE INDEX idx_progress_date ON progress (date);

CREATE INDEX idx_progress_lift_date ON progress (lift, date DESC);

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

CREATE TABLE routine (
    id TEXT PRIMARY KEY NOT NULL,
    steps TEXT NOT NULL,
    lift TEXT NOT NULL,
    FOREIGN KEY (lift) REFERENCES lift (id)
);

CREATE TABLE routine_workout_mapping (
    routine TEXT NOT NULL,
    workout TEXT NOT NULL,
    PRIMARY KEY (routine, workout),
    FOREIGN KEY (routine) REFERENCES routine (id),
    FOREIGN KEY (workout) REFERENCES workout (id)
);

CREATE TABLE workout (
    id TEXT PRIMARY KEY NOT NULL,
    template TEXT NOT NULL
);

CREATE TABLE subworkout (
    subworkout TEXT NOT NULL,
    superworkout TEXT NOT NULL,
    PRIMARY KEY (subworkout, superworkout),
    FOREIGN KEY (subworkout) REFERENCES workout (id),
    FOREIGN KEY (superworkout) REFERENCES workout (id)
);

CREATE TABLE lift_workout_mapping (
    lift TEXT NOT NULL,
    workout TEXT NOT NULL,
    PRIMARY KEY (lift, workout),
    FOREIGN KEY (lift) REFERENCES lift (id),
    FOREIGN KEY (workout) REFERENCES workout (id)
);

CREATE TABLE lift (
    id TEXT PRIMARY KEY NOT NULL,
    link TEXT NOT NULL,
    default_side_weight TEXT,
    notes TEXT,
    FOREIGN KEY (default_side_weight) REFERENCES side_weight (id)
);

CREATE TABLE lift_muscle_mapping (
    lift TEXT NOT NULL,
    muscle TEXT NOT NULL,
    movement TEXT NOT NULL,
    PRIMARY KEY (lift, muscle, movement),
    FOREIGN KEY (lift) REFERENCES lift (id),
    FOREIGN KEY (muscle) REFERENCES muscle (id),
    FOREIGN KEY (movement) REFERENCES movement (id)
);

CREATE TABLE muscle (
    id TEXT PRIMARY KEY NOT NULL,
    link TEXT NOT NULL,
    -- MRV/MEV/...
    message TEXT
);

CREATE TABLE movement (
    id TEXT PRIMARY KEY NOT NULL,
    alias TEXT NOT NULL
);

CREATE TABLE side_weight (
    id TEXT PRIMARY KEY NOT NULL,
    multiplier REAL NOT NULL,
    addend REAL NOT NULL,
    format TEXT NOT NULL
);

CREATE TABLE template_variable (
    id TEXT PRIMARY KEY NOT NULL,
    value TEXT NOT NULL
);

CREATE TABLE lift_groups (
    id TEXT PRIMARY KEY NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE progress;
DROP INDEX idx_progress_date;
DROP INDEX idx_progress_lift_date;
DROP TABLE schedule;
DROP TABLE schedule_list;
DROP TABLE routine;
DROP TABLE routine_workout_mapping;
DROP TABLE workout;
DROP TABLE subworkout;
DROP TABLE lift_workout_mapping;
DROP TABLE lift;
DROP TABLE lift_muscle_mapping;
DROP TABLE muscle;
DROP TABLE movement;
DROP TABLE side_weight;
DROP TABLE template_variable;
DROP TABLE lift_groups;
-- +goose StatementEnd

-- sqlfluff:dialect:sqlite
-- sqlfluff:rules:references.keywords:ignore_words:date
