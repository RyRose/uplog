package rawdata

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/service/rawdata/base"
	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

// HandleGetWorkoutView godoc
//
//	@Summary		Get workout data table view
//	@Description	Renders a paginated table view of workouts
//	@Tags			rawdata
//	@Produce		html
//	@Param			offset	query		integer	false	"Pagination offset"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/workout [get]
func HandleGetWorkoutView(roDB *sql.DB) http.HandlerFunc {
	return base.HandleGetDataTableView(
		roDB,
		base.TableViewMetadata{
			Headers: []string{"ID", "Template"},
			Post:    "/view/data/workout",
		},
		(*workoutdb.Queries).RawSelectWorkoutPage,
		func(limit, offset int64) workoutdb.RawSelectWorkoutPageParams {
			return workoutdb.RawSelectWorkoutPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, items []workoutdb.Workout) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/workout", item.ID),
					DeleteEndpoint: util.UrlPathJoin("/view/data/workout", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Type: templates.InputString, Value: item.ID},
						{Name: "template", Type: templates.TextArea, Value: item.Template},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "template", Type: templates.TextArea},
				},
			})
			return rows, nil
		},
	)
}

// HandlePatchWorkoutView godoc
//
//	@Summary		Update workout data
//	@Description	Updates specific fields of a workout entry by ID
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Param			id			path		string	true	"Workout ID"
//	@Param			id			formData	string	false	"New workout ID"
//	@Param			template	formData	string	false	"Workout template"
//	@Success		200			{string}	string	"OK"
//	@Failure		400			{string}	string	"Bad request"
//	@Failure		500			{string}	string	"Internal server error"
//	@Router			/view/data/workout/{id} [patch]
func HandlePatchWorkoutView(wDB *sql.DB) http.HandlerFunc {
	return base.HandlePatchTableRowViewID(
		wDB,
		map[string]base.PatcherID{
			"id": &base.PatchIDParams[workoutdb.RawUpdateWorkoutIdParams]{
				Query: (*workoutdb.Queries).RawUpdateWorkoutId,
				Convert: func(id, value string) (*workoutdb.RawUpdateWorkoutIdParams, error) {
					return &workoutdb.RawUpdateWorkoutIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"template": &base.PatchIDParams[workoutdb.RawUpdateWorkoutTemplateParams]{
				Query: (*workoutdb.Queries).RawUpdateWorkoutTemplate,
				Convert: func(id, value string) (*workoutdb.RawUpdateWorkoutTemplateParams, error) {
					return &workoutdb.RawUpdateWorkoutTemplateParams{
						ID:       id,
						Template: value,
					}, nil
				},
			},
		},
	)
}

// HandlePostWorkoutView godoc
//
//	@Summary		Create new workout
//	@Description	Creates a new workout entry in the database
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			id			formData	string	true	"Workout ID"
//	@Param			template	formData	string	true	"Workout template"
//	@Success		201			{string}	string	"HTML content"
//	@Failure		400			{string}	string	"Bad request"
//	@Failure		500			{string}	string	"Internal server error"
//	@Router			/view/data/workout [post]
func HandlePostWorkoutView(roDB, wDB *sql.DB) http.HandlerFunc {
	return base.HandlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertWorkout,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertWorkoutParams, error) {
			return &workoutdb.RawInsertWorkoutParams{
				ID:       values.Get("id"),
				Template: values.Get("template"),
			}, nil
		},
		func(_ context.Context, _ *workoutdb.Queries, item workoutdb.Workout) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/workout", item.ID),
				DeleteEndpoint: util.UrlPathJoin("/view/data/workout", item.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString, Value: item.ID},
					{Name: "template", Type: templates.TextArea, Value: item.Template},
				},
			}, nil
		},
	)
}

// HandleDeleteWorkoutView godoc
//
//	@Summary		Delete workout
//	@Description	Deletes a workout entry by ID
//	@Tags			rawdata
//	@Param			id	path		string	true	"Workout ID"
//	@Success		200	{string}	string	"OK"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/data/workout/{id} [delete]
func HandleDeleteWorkoutView(wDB *sql.DB) http.HandlerFunc {
	return base.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteWorkout)
}
