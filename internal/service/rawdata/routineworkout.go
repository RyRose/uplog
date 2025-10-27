package rawdata

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/config"
	"github.com/RyRose/uplog/internal/service/rawdata/base"
	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

// HandleGetRoutineWorkoutView godoc
//
//	@Summary		Get routine workout mapping data table view
//	@Description	Renders a paginated table view of routine-workout mappings
//	@Tags			rawdata
//	@Produce		html
//	@Param			offset	query		integer	false	"Pagination offset"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/routine_workout_mapping [get]
func HandleGetRoutineWorkoutView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleGetDataTableView(
		state.ReadonlyDB,
		base.TableViewMetadata{
			Headers: []string{"Routine", "Workout"},
			Post:    "/view/data/routine_workout_mapping",
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
					PatchEndpoint:  util.UrlPathJoin("/view/data/routine_workout_mapping", item.Routine, item.Workout),
					DeleteEndpoint: util.UrlPathJoin("/view/data/routine_workout_mapping", item.Routine, item.Workout),
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

// HandlePatchRoutineWorkoutView godoc
//
//	@Summary		Update routine workout mapping data
//	@Description	Updates specific fields of a routine-workout mapping
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Param			routine	path		string	true	"Routine ID"
//	@Param			workout	path		string	true	"Workout ID"
//	@Param			routine	formData	string	false	"New routine ID"
//	@Param			workout	formData	string	false	"New workout ID"
//	@Success		200		{string}	string	"OK"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/routine_workout_mapping/{routine}/{workout} [patch]
func HandlePatchRoutineWorkoutView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePatchTableRowViewRequest(
		state.WriteDB,
		map[string]base.PatcherReq{
			"routine": &base.PatchReqParams[workoutdb.RawUpdateRoutineWorkoutMappingRoutineParams]{
				Query: (*workoutdb.Queries).RawUpdateRoutineWorkoutMappingRoutine,
				Convert: func(r *http.Request, value string) (*workoutdb.RawUpdateRoutineWorkoutMappingRoutineParams, error) {
					return &workoutdb.RawUpdateRoutineWorkoutMappingRoutineParams{
						Out:     value,
						In:      r.PathValue("routine"),
						Workout: r.PathValue("workout"),
					}, nil
				},
			},
			"workout": &base.PatchReqParams[workoutdb.RawUpdateRoutineWorkoutMappingWorkoutParams]{
				Query: (*workoutdb.Queries).RawUpdateRoutineWorkoutMappingWorkout,
				Convert: func(r *http.Request, value string) (*workoutdb.RawUpdateRoutineWorkoutMappingWorkoutParams, error) {
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

// HandlePostRoutineWorkoutView godoc
//
//	@Summary		Create new routine workout mapping
//	@Description	Creates a new routine-workout mapping in the database
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			routine	formData	string	true	"Routine ID"
//	@Param			workout	formData	string	true	"Workout ID"
//	@Success		201		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/routine_workout_mapping [post]
func HandlePostRoutineWorkoutView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePostDataTableView(
		state.ReadonlyDB,
		state.WriteDB,
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
				PatchEndpoint:  util.UrlPathJoin("/view/data/routine_workout_mapping", item.Routine, item.Workout),
				DeleteEndpoint: util.UrlPathJoin("/view/data/routine_workout_mapping", item.Routine, item.Workout),
				Values: []templates.DataTableValue{
					{Name: "routine", Type: templates.Select, Value: item.Routine, SelectOptions: routines},
					{Name: "workout", Type: templates.Select, Value: item.Workout, SelectOptions: workouts},
				},
			}, nil
		},
	)
}

// HandleDeleteRoutineWorkoutView godoc
//
//	@Summary		Delete routine workout mapping
//	@Description	Deletes a routine-workout mapping
//	@Tags			rawdata
//	@Param			routine	path		string	true	"Routine ID"
//	@Param			workout	path		string	true	"Workout ID"
//	@Success		200		{string}	string	"OK"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/routine_workout_mapping/{routine}/{workout} [delete]
func HandleDeleteRoutineWorkoutView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleDeleteTableRowViewRequest(
		state.WriteDB,
		(*workoutdb.Queries).RawDeleteRoutineWorkout,
		func(r *http.Request) (*workoutdb.RawDeleteRoutineWorkoutParams, error) {
			return &workoutdb.RawDeleteRoutineWorkoutParams{
				Routine: r.PathValue("routine"),
				Workout: r.PathValue("workout"),
			}, nil
		},
	)
}
