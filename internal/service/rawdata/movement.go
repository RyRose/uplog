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

// HandleGetMovementView godoc
//
//	@Summary		Get movement data table view
//	@Description	Renders a paginated table view of movements
//	@Tags			rawdata
//	@Produce		html
//	@Param			offset	query		integer	false	"Pagination offset"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/movement [get]
func HandleGetMovementView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleGetDataTableView(
		state.RDB,
		base.TableViewMetadata{
			Headers: []string{"ID", "Alias"},
			Post:    "/view/data/movement",
		},
		(*workoutdb.Queries).RawSelectMovementPage,
		func(limit, offset int64) workoutdb.RawSelectMovementPageParams {
			return workoutdb.RawSelectMovementPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, movements []workoutdb.Movement) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, movement := range movements {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/movement", movement.ID),
					DeleteEndpoint: util.UrlPathJoin("/view/data/movement", movement.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Value: movement.ID, Type: templates.InputString},
						{Name: "alias", Value: movement.Alias, Type: templates.InputString},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "alias", Type: templates.InputString},
				},
			})
			return rows, nil
		},
	)
}

// HandlePatchMovementView godoc
//
//	@Summary		Update movement data
//	@Description	Updates specific fields of a movement entry by ID
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Param			id		path		string	true	"Movement ID"
//	@Param			id		formData	string	false	"New movement ID"
//	@Param			alias	formData	string	false	"Movement alias"
//	@Success		200		{string}	string	"OK"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/movement/{id} [patch]
func HandlePatchMovementView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePatchTableRowViewID(
		state.WDB,
		map[string]base.PatcherID{
			"id": &base.PatchIDParams[workoutdb.RawUpdateMovementIdParams]{
				Query: (*workoutdb.Queries).RawUpdateMovementId,
				Convert: func(id, value string) (*workoutdb.RawUpdateMovementIdParams, error) {
					return &workoutdb.RawUpdateMovementIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"alias": &base.PatchIDParams[workoutdb.RawUpdateMovementAliasParams]{
				Query: (*workoutdb.Queries).RawUpdateMovementAlias,
				Convert: func(id, value string) (*workoutdb.RawUpdateMovementAliasParams, error) {
					return &workoutdb.RawUpdateMovementAliasParams{
						ID:    id,
						Alias: value,
					}, nil
				},
			},
		},
	)
}

// HandlePostMovementView godoc
//
//	@Summary		Create new movement
//	@Description	Creates a new movement entry in the database
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			id		formData	string	true	"Movement ID"
//	@Param			alias	formData	string	true	"Movement alias"
//	@Success		201		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/movement [post]
func HandlePostMovementView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePostDataTableView(
		state.RDB,
		state.WDB,
		(*workoutdb.Queries).RawInsertMovement,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertMovementParams, error) {
			return &workoutdb.RawInsertMovementParams{
				ID:    values.Get("id"),
				Alias: values.Get("alias"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, movement workoutdb.Movement) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/movement", movement.ID),
				DeleteEndpoint: util.UrlPathJoin("/view/data/movement", movement.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Value: movement.ID, Type: templates.InputString},
					{Name: "alias", Value: movement.Alias, Type: templates.InputString},
				},
			}, nil
		},
	)
}

// HandleDeleteMovementView godoc
//
//	@Summary		Delete movement
//	@Description	Deletes a movement entry by ID
//	@Tags			rawdata
//	@Param			id	path		string	true	"Movement ID"
//	@Success		200	{string}	string	"OK"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/data/movement/{id} [delete]
func HandleDeleteMovementView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleDeleteTableRowViewID(state.WDB, (*workoutdb.Queries).RawDeleteMovement)
}
