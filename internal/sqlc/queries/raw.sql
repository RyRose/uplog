-- name: RawInsertProgress :one
INSERT INTO progress (lift, date, weight, sets, reps, side_weight)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: RawSelectProgress :many
SELECT * FROM progress;

-- name: RawSelectProgressPage :many
SELECT * FROM progress LIMIT ? OFFSET ?;

-- name: RawDeleteProgress :exec
DELETE FROM progress
WHERE id = ?;

-- name: RawUpdateProgressLift :exec
UPDATE progress
SET lift = ?
WHERE id = ?;

-- name: RawUpdateProgressDate :exec
UPDATE progress
SET date = ?
WHERE id = ?;

-- name: RawUpdateProgressWeight :exec
UPDATE progress
SET weight = ?
WHERE id = ?;

-- name: RawUpdateProgressSets :exec
UPDATE progress
SET sets = ?
WHERE id = ?;

-- name: RawUpdateProgressReps :exec
UPDATE progress
SET reps = ?
WHERE id = ?;

-- name: RawUpdateProgressSideWeight :exec
UPDATE progress
SET side_weight = ?
WHERE id = ?;

-- name: RawInsertRoutine :one
INSERT INTO routine (id, steps, lift)
VALUES (?, ?, ?)
RETURNING *;

-- name: RawSelectRoutine :many
SELECT * FROM routine;

-- name: RawSelectRoutinePage :many
SELECT * FROM routine LIMIT ? OFFSET ?;

-- name: RawDeleteRoutine :exec
DELETE FROM routine
WHERE id = ?;

-- name: RawUpdateRoutineId :exec
UPDATE routine
SET id = @out
WHERE id = @in;

-- name: RawUpdateRoutineSteps :exec
UPDATE routine
SET steps = ?
WHERE id = ?;

-- name: RawUpdateRoutineLift :exec
UPDATE routine
SET lift = ?
WHERE id = ?;

-- name: RawInsertRoutineWorkout :one
INSERT INTO routine_workout_mapping (routine, workout)
VALUES (?, ?)
RETURNING *;

-- name: RawSelectRoutineWorkout :many
SELECT * FROM routine_workout_mapping;

-- name: RawSelectRoutineWorkoutPage :many
SELECT * FROM routine_workout_mapping LIMIT ? OFFSET ?;

-- name: RawDeleteRoutineWorkout :exec
DELETE FROM routine_workout_mapping
WHERE routine = ? AND workout = ?;

-- name: RawUpdateRoutineWorkoutMappingRoutine :exec
UPDATE routine_workout_mapping
SET routine = @out
WHERE routine = @in AND workout = ?;

-- name: RawUpdateRoutineWorkoutMappingWorkout :exec
UPDATE routine_workout_mapping
SET workout = @out
WHERE routine = ? AND workout = @in;

-- name: RawInsertWorkout :one
INSERT INTO workout (id, template)
VALUES (?, ?)
RETURNING *;

-- name: RawSelectWorkout :many
SELECT * FROM workout;

-- name: RawSelectWorkoutPage :many
SELECT * FROM workout LIMIT ? OFFSET ?;

-- name: RawDeleteWorkout :exec
DELETE FROM workout
WHERE id = ?;

-- name: RawUpdateWorkoutId :exec
UPDATE workout
SET id = @out
WHERE id = @in;

-- name: RawUpdateWorkoutTemplate :exec
UPDATE workout
SET template = ?
WHERE id = ?;

-- name: RawInsertSubworkout :one
INSERT INTO subworkout (subworkout, superworkout)
VALUES (?, ?)
RETURNING *;

-- name: RawSelectSubworkout :many
SELECT * FROM subworkout;

-- name: RawSelectSubworkoutPage :many
SELECT * FROM subworkout LIMIT ? OFFSET ?;

-- name: RawDeleteSubworkout :exec
DELETE FROM subworkout
WHERE subworkout = ? AND superworkout = ?;

-- name: RawUpdateSubworkoutSubworkout :exec
UPDATE subworkout
SET subworkout = @out
WHERE superworkout = ? AND subworkout = @in;

-- name: RawUpdateSubworkoutSuperworkout :exec
UPDATE subworkout
SET superworkout = @out
WHERE superworkout = @in AND subworkout = ?;

-- name: RawInsertLiftWorkout :one
INSERT INTO lift_workout_mapping (lift, workout)
VALUES (?, ?)
RETURNING *;

-- name: RawSelectLiftWorkout :many
SELECT * FROM lift_workout_mapping;

-- name: RawSelectLiftWorkoutPage :many
SELECT * FROM lift_workout_mapping LIMIT ? OFFSET ?;

-- name: RawDeleteLiftWorkout :exec
DELETE FROM lift_workout_mapping
WHERE lift = ? AND workout = ?;

-- name: RawUpdateLiftWorkoutMappingLift :exec
UPDATE lift_workout_mapping
SET lift = @out
WHERE lift = @in AND workout = ?;

-- name: RawUpdateLiftWorkoutMappingWorkout :exec
UPDATE lift_workout_mapping
SET workout = @out
WHERE lift = ? AND workout = @in;

-- name: RawInsertLift :one
INSERT INTO lift (id, link, default_side_weight, notes, lift_group)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: RawSelectLift :many
SELECT * FROM lift;

-- name: RawSelectLiftPage :many
SELECT * FROM lift LIMIT ? OFFSET ?;

-- name: RawDeleteLift :exec
DELETE FROM lift
WHERE id = ?;

-- name: RawUpdateLiftId :exec
UPDATE lift
SET id = @out
WHERE id = @in;

-- name: RawUpdateLiftLink :exec
UPDATE lift
SET link = ?
WHERE id = ?;

-- name: RawUpdateLiftDefaultSideWeight :exec
UPDATE lift
SET default_side_weight = ?
WHERE id = ?;

-- name: RawUpdateLiftNotes :exec
UPDATE lift
SET notes = ?
WHERE id = ?;

-- name: RawUpdateLiftLiftGroup :exec
UPDATE lift
SET lift_group = ?
WHERE id = ?;

-- name: RawInsertLiftMuscle :one
INSERT INTO lift_muscle_mapping (lift, muscle, movement)
VALUES (?, ?, ?)
RETURNING *;

-- name: RawSelectLiftMuscle :many
SELECT * FROM lift_muscle_mapping;

-- name: RawSelectLiftMusclePage :many
SELECT * FROM lift_muscle_mapping LIMIT ? OFFSET ?;

-- name: RawDeleteLiftMuscle :exec
DELETE FROM lift_muscle_mapping
WHERE lift = ? AND muscle = ? AND movement = ?;

-- name: RawUpdateLiftMuscleMappingLift :exec
UPDATE lift_muscle_mapping
SET lift = @out
WHERE lift = @in AND muscle = ? AND movement = ?;

-- name: RawUpdateLiftMuscleMappingMuscle :exec
UPDATE lift_muscle_mapping
SET muscle = @out
WHERE lift = ? AND muscle = @in AND movement = ?;

-- name: RawUpdateLiftMuscleMappingMovement :exec
UPDATE lift_muscle_mapping
SET movement = @out
WHERE lift = ? AND muscle = ? AND movement = @in;

-- name: RawInsertMuscle :one
INSERT INTO muscle (id, link, message)
VALUES (?, ?, ?)
RETURNING *;

-- name: RawSelectMuscle :many
SELECT * FROM muscle;

-- name: RawSelectMusclePage :many
SELECT * FROM muscle LIMIT ? OFFSET ?;

-- name: RawDeleteMuscle :exec
DELETE FROM muscle
WHERE id = ?;

-- name: RawUpdateMuscleId :exec
UPDATE muscle
SET id = @out
WHERE id = @in;

-- name: RawUpdateMuscleLink :exec
UPDATE muscle
SET link = ?
WHERE id = ?;

-- name: RawUpdateMuscleMessage :exec
UPDATE muscle
SET message = ?
WHERE id = ?;

-- name: RawInsertMovement :one
INSERT INTO movement (id, alias)
VALUES (?, ?)
RETURNING *;

-- name: RawSelectMovement :many
SELECT * FROM movement;

-- name: RawSelectMovementPage :many
SELECT * FROM movement LIMIT ? OFFSET ?;

-- name: RawDeleteMovement :exec
DELETE FROM movement
WHERE id = ?;

-- name: RawUpdateMovementId :exec
UPDATE movement
SET id = @out
WHERE id = @in;

-- name: RawUpdateMovementAlias :exec
UPDATE movement
SET alias = ?
WHERE id = ?;

-- name: RawInsertSideWeight :one
INSERT INTO side_weight (id, multiplier, addend, format)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: RawSelectSideWeight :many
SELECT * FROM side_weight;

-- name: RawSelectSideWeightPage :many
SELECT * FROM side_weight LIMIT ? OFFSET ?;

-- name: RawDeleteSideWeight :exec
DELETE FROM side_weight
WHERE id = ?;

-- name: RawUpdateSideWeightId :exec
UPDATE side_weight
SET id = @out
WHERE id = @in;

-- name: RawUpdateSideWeightMultiplier :exec
UPDATE side_weight
SET multiplier = ?
WHERE id = ?;

-- name: RawUpdateSideWeightAddend :exec
UPDATE side_weight
SET addend = ?
WHERE id = ?;

-- name: RawUpdateSideWeightFormat :exec
UPDATE side_weight
SET format = ?
WHERE id = ?;

-- name: RawInsertTemplateVariable :one
INSERT INTO template_variable (id, value)
VALUES (?, ?)
RETURNING *;

-- name: RawSelectTemplateVariable :many
SELECT * FROM template_variable;

-- name: RawSelectTemplateVariablePage :many
SELECT * FROM template_variable LIMIT ? OFFSET ?;

-- name: RawDeleteTemplateVariable :exec
DELETE FROM template_variable
WHERE id = ?;

-- name: RawUpdateTemplateVariableId :exec
UPDATE template_variable
SET id = @out
WHERE id = @in;

-- name: RawUpdateTemplateVariableValue :exec
UPDATE template_variable
SET value = ?
WHERE id = ?;

-- name: RawInsertLiftGroup :one
INSERT INTO lift_group (id)
VALUES (?)
RETURNING *;

-- name: RawSelectLiftGroup :many
SELECT * FROM lift_group;

-- name: RawSelectLiftGroupPage :many
SELECT * FROM lift_group LIMIT ? OFFSET ?;

-- name: RawDeleteLiftGroup :exec
DELETE FROM lift_group
WHERE id = ?;

-- name: RawUpdateLiftGroupId :exec
UPDATE lift_group
SET id = @out
WHERE id = @in;

-----------------------
-- sqlfluff settings --
-----------------------

-- Treat as sqlite.
-- sqlfluff:dialect:sqlite

-- Disable unknown column count rule. The structs generated by `*`
-- are better and are handled gracefully by sqlc.
-- sqlfluff:exclude_rules:L044
