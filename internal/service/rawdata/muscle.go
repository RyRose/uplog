package rawdata

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/config"
	"github.com/RyRose/uplog/internal/service/rawdata/base"
	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

// HandleGetMuscleView godoc
//
//	@Summary		Get muscle data table view
//	@Description	Renders a paginated table view of muscles
//	@Tags			rawdata
//	@Produce		html
//	@Param			offset	query		integer	false	"Pagination offset"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/muscle [get]
func HandleGetMuscleView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleGetDataTableView(
		state.ReadonlyDB,
		base.TableViewMetadata{
			Headers: []string{"ID", "Link", "Message"},
			Post:    "/view/data/muscle",
		},
		(*workoutdb.Queries).RawSelectMusclePage,
		func(limit, offset int64) workoutdb.RawSelectMusclePageParams {
			return workoutdb.RawSelectMusclePageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, items []workoutdb.Muscle) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/muscle", item.ID),
					DeleteEndpoint: util.UrlPathJoin("/view/data/muscle", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Value: item.ID, Type: templates.InputString},
						{Name: "link", Value: item.Link, Type: templates.InputString},
						{Name: "message", Value: util.Zero(item.Message), Type: templates.InputString},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "link", Type: templates.InputString},
					{Name: "muscle", Type: templates.InputString},
				},
			})
			return rows, nil
		},
	)
}

// HandlePatchMuscleView godoc
//
//	@Summary		Update muscle data
//	@Description	Updates specific fields of a muscle entry by ID
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Param			id		path		string	true	"Muscle ID"
//	@Param			id		formData	string	false	"New muscle ID"
//	@Param			link	formData	string	false	"Muscle link"
//	@Param			message	formData	string	false	"Message"
//	@Success		200		{string}	string	"OK"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/muscle/{id} [patch]
func HandlePatchMuscleView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePatchTableRowViewID(
		state.WriteDB,
		map[string]base.PatcherID{
			"id": &base.PatchIDParams[workoutdb.RawUpdateMuscleIdParams]{
				Query: (*workoutdb.Queries).RawUpdateMuscleId,
				Convert: func(id, value string) (*workoutdb.RawUpdateMuscleIdParams, error) {
					return &workoutdb.RawUpdateMuscleIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"link": &base.PatchIDParams[workoutdb.RawUpdateMuscleLinkParams]{
				Query: (*workoutdb.Queries).RawUpdateMuscleLink,
				Convert: func(id, value string) (*workoutdb.RawUpdateMuscleLinkParams, error) {
					return &workoutdb.RawUpdateMuscleLinkParams{
						ID:   id,
						Link: value,
					}, nil
				},
			},
			"message": &base.PatchIDParams[workoutdb.RawUpdateMuscleMessageParams]{
				Query: (*workoutdb.Queries).RawUpdateMuscleMessage,
				Convert: func(id, value string) (*workoutdb.RawUpdateMuscleMessageParams, error) {
					return &workoutdb.RawUpdateMuscleMessageParams{
						ID:      id,
						Message: util.DeZero(value),
					}, nil
				},
			},
		},
	)
}

// HandlePostMuscleView godoc
//
//	@Summary		Create new muscle
//	@Description	Creates a new muscle entry in the database
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			id		formData	string	true	"Muscle ID"
//	@Param			link	formData	string	true	"Muscle link"
//	@Param			message	formData	string	false	"Message"
//	@Success		201		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/muscle [post]
func HandlePostMuscleView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePostDataTableView(
		state.ReadonlyDB,
		state.WriteDB,
		(*workoutdb.Queries).RawInsertMuscle,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertMuscleParams, error) {
			return &workoutdb.RawInsertMuscleParams{
				ID:      values.Get("id"),
				Link:    values.Get("link"),
				Message: util.DeZero(values.Get("message")),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, movement workoutdb.Muscle) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/muscle", movement.ID),
				DeleteEndpoint: util.UrlPathJoin("/view/data/muscle", movement.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: movement.ID, Type: templates.InputString},
					{Name: "link", Value: movement.Link, Type: templates.InputString},
					{Name: "message", Value: util.Zero(movement.Message), Type: templates.InputString},
				},
			}, nil
		},
	)
}

// HandleDeleteMuscleView godoc
//
//	@Summary		Delete muscle
//	@Description	Deletes a muscle entry by ID
//	@Tags			rawdata
//	@Param			id	path		string	true	"Muscle ID"
//	@Success		200	{string}	string	"OK"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/data/muscle/{id} [delete]
func HandleDeleteMuscleView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleDeleteTableRowViewID(state.WriteDB, (*workoutdb.Queries).RawDeleteMuscle)
}
