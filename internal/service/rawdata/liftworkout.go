package rawdata

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/service/rawdata/base"
	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

// HandleGetLiftWorkoutView godoc
//
//	@Summary		Get lift workout mapping data table view
//	@Description	Renders a paginated table view of lift-workout mappings
//	@Tags			rawdata
//	@Produce		html
//	@Param			offset	query		integer	false	"Pagination offset"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/lift_workout_mapping [get]
func HandleGetLiftWorkoutView(roDB *sql.DB) http.HandlerFunc {
	return base.HandleGetDataTableView(
		roDB,
		base.TableViewMetadata{
			Headers: []string{"Lift", "Workout"},
			Post:    "/view/data/lift_workout_mapping",
		},
		(*workoutdb.Queries).RawSelectLiftWorkoutPage,
		func(limit, offset int64) workoutdb.RawSelectLiftWorkoutPageParams {
			return workoutdb.RawSelectLiftWorkoutPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.LiftWorkoutMapping) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/lift_workout_mapping", item.Lift, item.Workout),
					DeleteEndpoint: util.UrlPathJoin("/view/data/lift_workout_mapping", item.Lift, item.Workout),
					Values: []templates.DataTableValue{
						{Name: "lift", Type: templates.Select, Value: item.Lift, SelectOptions: lifts},
						{Name: "workout", Type: templates.Select, Value: item.Workout, SelectOptions: workouts},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "lift", Type: templates.Select, SelectOptions: lifts},
					{Name: "workout", Type: templates.Select, SelectOptions: workouts},
				},
			})
			return rows, nil
		},
	)
}

// HandlePatchLiftWorkoutView godoc
//
//	@Summary		Update lift workout mapping data
//	@Description	Updates specific fields of a lift-workout mapping
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Param			lift	path		string	true	"Lift ID"
//	@Param			workout	path		string	true	"Workout ID"
//	@Param			lift	formData	string	false	"New lift ID"
//	@Param			workout	formData	string	false	"New workout ID"
//	@Success		200		{string}	string	"OK"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/lift_workout_mapping/{lift}/{workout} [patch]
func HandlePatchLiftWorkoutView(wDB *sql.DB) http.HandlerFunc {
	return base.HandlePatchTableRowViewRequest(
		wDB,
		map[string]base.PatcherReq{
			"lift": &base.PatchReqParams[workoutdb.RawUpdateLiftWorkoutMappingLiftParams]{
				Query: (*workoutdb.Queries).RawUpdateLiftWorkoutMappingLift,
				Convert: func(r *http.Request, value string) (*workoutdb.RawUpdateLiftWorkoutMappingLiftParams, error) {
					return &workoutdb.RawUpdateLiftWorkoutMappingLiftParams{
						Out:     value,
						In:      r.PathValue("lift"),
						Workout: r.PathValue("workout"),
					}, nil
				},
			},
			"workout": &base.PatchReqParams[workoutdb.RawUpdateLiftWorkoutMappingWorkoutParams]{
				Query: (*workoutdb.Queries).RawUpdateLiftWorkoutMappingWorkout,
				Convert: func(r *http.Request, value string) (*workoutdb.RawUpdateLiftWorkoutMappingWorkoutParams, error) {
					return &workoutdb.RawUpdateLiftWorkoutMappingWorkoutParams{
						Out:  value,
						In:   r.PathValue("workout"),
						Lift: r.PathValue("lift"),
					}, nil
				},
			},
		},
	)
}

// HandlePostLiftWorkoutView godoc
//
//	@Summary		Create new lift workout mapping
//	@Description	Creates a new lift-workout mapping in the database
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			lift	formData	string	true	"Lift ID"
//	@Param			workout	formData	string	true	"Workout ID"
//	@Success		201		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/lift_workout_mapping [post]
func HandlePostLiftWorkoutView(roDB, wDB *sql.DB) http.HandlerFunc {
	return base.HandlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertLiftWorkout,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertLiftWorkoutParams, error) {
			return &workoutdb.RawInsertLiftWorkoutParams{
				Lift:    values.Get("lift"),
				Workout: values.Get("workout"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.LiftWorkoutMapping) (*templates.DataTableRow, error) {
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}

			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/lift_workout_mapping", item.Lift, item.Workout),
				DeleteEndpoint: util.UrlPathJoin("/view/data/lift_workout_mapping", item.Lift, item.Workout),
				Values: []templates.DataTableValue{
					{Name: "lift", Type: templates.Select, Value: item.Lift, SelectOptions: lifts},
					{Name: "workout", Type: templates.Select, Value: item.Workout, SelectOptions: workouts},
				},
			}, nil
		},
	)
}

// HandleDeleteLiftWorkoutView godoc
//
//	@Summary		Delete lift workout mapping
//	@Description	Deletes a lift-workout mapping
//	@Tags			rawdata
//	@Param			lift	path		string	true	"Lift ID"
//	@Param			workout	path		string	true	"Workout ID"
//	@Success		200		{string}	string	"OK"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/lift_workout_mapping/{lift}/{workout} [delete]
func HandleDeleteLiftWorkoutView(wDB *sql.DB) http.HandlerFunc {
	return base.HandleDeleteTableRowViewRequest(
		wDB,
		(*workoutdb.Queries).RawDeleteLiftWorkout,
		func(r *http.Request) (*workoutdb.RawDeleteLiftWorkoutParams, error) {
			return &workoutdb.RawDeleteLiftWorkoutParams{
				Lift:    r.PathValue("lift"),
				Workout: r.PathValue("workout"),
			}, nil
		},
	)
}
