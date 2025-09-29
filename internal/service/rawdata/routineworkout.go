package rawdata

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetRoutineWorkoutView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"Routine", "Workout"},
			post:    "/view/data/routine_workout_mapping",
		},
		(*workoutdb.Queries).RawSelectRoutineWorkoutPage,
		func(limit, offset int64) workoutdb.RawSelectRoutineWorkoutPageParams {
			return workoutdb.RawSelectRoutineWorkoutPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.RoutineWorkoutMapping) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			routines, err := q.ListAllIndividualRoutines(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list routines: %w", err)
			}
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  urlPathJoin("/view/data/routine_workout_mapping", item.Routine, item.Workout),
					DeleteEndpoint: urlPathJoin("/view/data/routine_workout_mapping", item.Routine, item.Workout),
					Values: []templates.DataTableValue{
						{Name: "routine", Type: templates.Select, Value: item.Routine, SelectOptions: routines},
						{Name: "workout", Type: templates.Select, Value: item.Workout, SelectOptions: workouts},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "routine", Type: templates.Select, SelectOptions: routines},
					{Name: "workout", Type: templates.Select, SelectOptions: workouts},
				},
			})
			return rows, nil
		},
	)
}

func HandlePatchRoutineWorkoutView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewRequest(
		wDB,
		map[string]patcherReq{
			"routine": &patchReqParams[workoutdb.RawUpdateRoutineWorkoutMappingRoutineParams]{
				query: (*workoutdb.Queries).RawUpdateRoutineWorkoutMappingRoutine,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateRoutineWorkoutMappingRoutineParams, error) {
					return &workoutdb.RawUpdateRoutineWorkoutMappingRoutineParams{
						Out:     value,
						In:      r.PathValue("routine"),
						Workout: r.PathValue("workout"),
					}, nil
				},
			},
			"workout": &patchReqParams[workoutdb.RawUpdateRoutineWorkoutMappingWorkoutParams]{
				query: (*workoutdb.Queries).RawUpdateRoutineWorkoutMappingWorkout,
				convert: func(r *http.Request, value string) (*workoutdb.RawUpdateRoutineWorkoutMappingWorkoutParams, error) {
					return &workoutdb.RawUpdateRoutineWorkoutMappingWorkoutParams{
						Out:     value,
						In:      r.PathValue("workout"),
						Routine: r.PathValue("routine"),
					}, nil
				},
			},
		},
	)
}

func HandlePostRoutineWorkoutView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertRoutineWorkout,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertRoutineWorkoutParams, error) {
			return &workoutdb.RawInsertRoutineWorkoutParams{
				Routine: values.Get("routine"),
				Workout: values.Get("workout"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.RoutineWorkoutMapping) (*templates.DataTableRow, error) {
			routines, err := q.ListAllIndividualRoutines(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list routines: %w", err)
			}
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}

			return &templates.DataTableRow{
				PatchEndpoint:  urlPathJoin("/view/data/routine_workout_mapping", item.Routine, item.Workout),
				DeleteEndpoint: urlPathJoin("/view/data/routine_workout_mapping", item.Routine, item.Workout),
				Values: []templates.DataTableValue{
					{Name: "routine", Type: templates.Select, Value: item.Routine, SelectOptions: routines},
					{Name: "workout", Type: templates.Select, Value: item.Workout, SelectOptions: workouts},
				},
			}, nil
		},
	)
}
