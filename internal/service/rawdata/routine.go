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

// HandleGetRoutineView godoc
//
//	@Summary		Get routine data table view
//	@Description	Renders a paginated table view of routines
//	@Tags			rawdata
//	@Produce		html
//	@Param			offset	query		integer	false	"Pagination offset"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/routine [get]
func HandleGetRoutineView(roDB *sql.DB) http.HandlerFunc {
	return base.HandleGetDataTableView(
		roDB,
		base.TableViewMetadata{
			Headers: []string{"ID", "Steps", "Lift"},
			Post:    "/view/data/routine",
		},
		(*workoutdb.Queries).RawSelectRoutinePage,
		func(limit, offset int64) workoutdb.RawSelectRoutinePageParams {
			return workoutdb.RawSelectRoutinePageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.Routine) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}

			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/routine", item.ID),
					DeleteEndpoint: util.UrlPathJoin("/view/data/routine", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Value: item.ID, Type: templates.InputString},
						{Name: "steps", Value: item.Steps, Type: templates.InputString},
						{Name: "lift", Value: item.Lift, Type: templates.Select, SelectOptions: lifts},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "link", Type: templates.InputString},
					{Name: "lift", Type: templates.Select, SelectOptions: lifts},
				},
			})
			return rows, nil
		},
	)
}

// HandlePatchRoutineView godoc
//
//	@Summary		Update routine data
//	@Description	Updates specific fields of a routine entry by ID
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Param			id		path		string	true	"Routine ID"
//	@Param			id		formData	string	false	"New routine ID"
//	@Param			steps	formData	string	false	"Routine steps"
//	@Param			lift	formData	string	false	"Lift ID"
//	@Success		200		{string}	string	"OK"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/routine/{id} [patch]
func HandlePatchRoutineView(wDB *sql.DB) http.HandlerFunc {
	return base.HandlePatchTableRowViewID(
		wDB,
		map[string]base.PatcherID{
			"id": &base.PatchIDParams[workoutdb.RawUpdateRoutineIdParams]{
				Query: (*workoutdb.Queries).RawUpdateRoutineId,
				Convert: func(id, value string) (*workoutdb.RawUpdateRoutineIdParams, error) {
					return &workoutdb.RawUpdateRoutineIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"steps": &base.PatchIDParams[workoutdb.RawUpdateRoutineStepsParams]{
				Query: (*workoutdb.Queries).RawUpdateRoutineSteps,
				Convert: func(id, value string) (*workoutdb.RawUpdateRoutineStepsParams, error) {
					return &workoutdb.RawUpdateRoutineStepsParams{
						ID:    id,
						Steps: value,
					}, nil
				},
			},
			"lift": &base.PatchIDParams[workoutdb.RawUpdateRoutineLiftParams]{
				Query: (*workoutdb.Queries).RawUpdateRoutineLift,
				Convert: func(id, value string) (*workoutdb.RawUpdateRoutineLiftParams, error) {
					return &workoutdb.RawUpdateRoutineLiftParams{
						ID:   id,
						Lift: value,
					}, nil
				},
			},
		},
	)
}

// HandlePostRoutineView godoc
//
//	@Summary		Create new routine
//	@Description	Creates a new routine entry in the database
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			id		formData	string	true	"Routine ID"
//	@Param			steps	formData	string	true	"Routine steps"
//	@Param			lift	formData	string	true	"Lift ID"
//	@Success		201		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/routine [post]
func HandlePostRoutineView(roDB, wDB *sql.DB) http.HandlerFunc {
	return base.HandlePostDataTableView(
		roDB,
		wDB,
		(*workoutdb.Queries).RawInsertRoutine,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertRoutineParams, error) {
			return &workoutdb.RawInsertRoutineParams{
				ID:    values.Get("id"),
				Steps: values.Get("steps"),
				Lift:  values.Get("lift"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.Routine) (*templates.DataTableRow, error) {
			lifts, err := q.ListAllIndividualLifts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list lifts: %w", err)
			}

			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/routine", item.ID),
				DeleteEndpoint: util.UrlPathJoin("/view/data/routine", item.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: item.ID, Type: templates.InputString},
					{Name: "steps", Value: item.Steps, Type: templates.InputString},
					{Name: "lift", Value: item.Lift, Type: templates.Select, SelectOptions: lifts},
				},
			}, nil
		},
	)
}

// HandleDeleteRoutineView godoc
//
//	@Summary		Delete routine
//	@Description	Deletes a routine entry by ID
//	@Tags			rawdata
//	@Param			id	path		string	true	"Routine ID"
//	@Success		200	{string}	string	"OK"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/data/routine/{id} [delete]
func HandleDeleteRoutineView(wDB *sql.DB) http.HandlerFunc {
	return base.HandleDeleteTableRowViewID(wDB, (*workoutdb.Queries).RawDeleteRoutine)
}
