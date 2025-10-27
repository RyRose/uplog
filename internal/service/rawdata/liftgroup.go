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

// HandleGetLiftGroupView godoc
//
//	@Summary		Get lift group data table view
//	@Description	Renders a paginated table view of lift groups
//	@Tags			rawdata
//	@Produce		html
//	@Param			offset	query		integer	false	"Pagination offset"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/lift_group [get]
func HandleGetLiftGroupView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleGetDataTableView(
		state.ReadonlyDB,
		base.TableViewMetadata{
			Headers: []string{"ID"},
			Post:    "/view/data/lift_group",
		},
		(*workoutdb.Queries).RawSelectLiftGroupPage,
		func(limit, offset int64) workoutdb.RawSelectLiftGroupPageParams {
			return workoutdb.RawSelectLiftGroupPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, liftGroups []string) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, liftGroup := range liftGroups {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/lift_group", liftGroup),
					DeleteEndpoint: util.UrlPathJoin("/view/data/lift_group", liftGroup),
					Values: []templates.DataTableValue{
						{Name: "id", Value: liftGroup, Type: templates.InputString},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
				},
			})
			return rows, nil
		},
	)
}

// HandlePatchLiftGroupView godoc
//
//	@Summary		Update lift group data
//	@Description	Updates the ID of a lift group entry
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Param			id	path		string	true	"Lift group ID"
//	@Param			id	formData	string	false	"New lift group ID"
//	@Success		200	{string}	string	"OK"
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/data/lift_group/{id} [patch]
func HandlePatchLiftGroupView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePatchTableRowViewID(
		state.WriteDB,
		map[string]base.PatcherID{
			"id": &base.PatchIDParams[workoutdb.RawUpdateLiftGroupIdParams]{
				Query: (*workoutdb.Queries).RawUpdateLiftGroupId,
				Convert: func(id, value string) (*workoutdb.RawUpdateLiftGroupIdParams, error) {
					return &workoutdb.RawUpdateLiftGroupIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
		},
	)
}

// HandlePostLiftGroupView godoc
//
//	@Summary		Create new lift group
//	@Description	Creates a new lift group entry in the database
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			id	formData	string	true	"Lift group ID"
//	@Success		201	{string}	string	"HTML content"
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/data/lift_group [post]
func HandlePostLiftGroupView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePostDataTableView(
		state.ReadonlyDB,
		state.WriteDB,
		(*workoutdb.Queries).RawInsertLiftGroup,
		func(_ context.Context, values url.Values) (*string, error) {
			id := values.Get("id")
			return &id, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, liftGroup string) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/lift_group", liftGroup),
				DeleteEndpoint: util.UrlPathJoin("/view/data/lift_group", liftGroup),
				Values: []templates.DataTableValue{
					{Name: "id", Value: liftGroup, Type: templates.InputString},
				},
			}, nil
		},
	)
}

// HandleDeleteLiftGroupView godoc
//
//	@Summary		Delete lift group
//	@Description	Deletes a lift group entry by ID
//	@Tags			rawdata
//	@Param			id	path		string	true	"Lift group ID"
//	@Success		200	{string}	string	"OK"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/data/lift_group/{id} [delete]
func HandleDeleteLiftGroupView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleDeleteTableRowViewID(state.WriteDB, (*workoutdb.Queries).RawDeleteLiftGroup)
}
