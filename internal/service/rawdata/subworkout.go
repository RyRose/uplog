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

// HandleGetSubworkoutView godoc
//
//	@Summary		Get subworkout data table view
//	@Description	Renders a paginated table view of subworkout-superworkout relationships
//	@Tags			rawdata
//	@Produce		html
//	@Param			offset	query		integer	false	"Pagination offset"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/subworkout [get]
func HandleGetSubworkoutView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleGetDataTableView(
		state.ReadonlyDB,
		base.TableViewMetadata{
			Headers: []string{"Subworkout", "Superworkout"},
			Post:    "/view/data/subworkout",
		},
		(*workoutdb.Queries).RawSelectSubworkoutPage,
		func(limit, offset int64) workoutdb.RawSelectSubworkoutPageParams {
			return workoutdb.RawSelectSubworkoutPageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(ctx context.Context, roDB *sql.DB, items []workoutdb.Subworkout) ([]templates.DataTableRow, error) {
			q := workoutdb.New(roDB)
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/subworkout", item.Subworkout, item.Superworkout),
					DeleteEndpoint: util.UrlPathJoin("/view/data/subworkout", item.Subworkout, item.Superworkout),
					Values: []templates.DataTableValue{
						{Name: "subworkout", Type: templates.Select, Value: item.Subworkout, SelectOptions: workouts},
						{Name: "superworkout", Type: templates.Select, Value: item.Superworkout, SelectOptions: workouts},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "subworkout", Type: templates.Select, SelectOptions: workouts},
					{Name: "superworkout", Type: templates.Select, SelectOptions: workouts},
				},
			})
			return rows, nil
		},
	)
}

// HandlePatchSubworkoutView godoc
//
//	@Summary		Update subworkout data
//	@Description	Updates specific fields of a subworkout-superworkout relationship
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Param			subworkout		path		string	true	"Subworkout ID"
//	@Param			superworkout	path		string	true	"Superworkout ID"
//	@Param			subworkout		formData	string	false	"New subworkout ID"
//	@Param			superworkout	formData	string	false	"New superworkout ID"
//	@Success		200				{string}	string	"OK"
//	@Failure		400				{string}	string	"Bad request"
//	@Failure		500				{string}	string	"Internal server error"
//	@Router			/view/data/subworkout/{subworkout}/{superworkout} [patch]
func HandlePatchSubworkoutView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePatchTableRowViewRequest(
		state.WriteDB,
		map[string]base.PatcherReq{
			"subworkout": &base.PatchReqParams[workoutdb.RawUpdateSubworkoutSubworkoutParams]{
				Query: (*workoutdb.Queries).RawUpdateSubworkoutSubworkout,
				Convert: func(r *http.Request, value string) (*workoutdb.RawUpdateSubworkoutSubworkoutParams, error) {
					return &workoutdb.RawUpdateSubworkoutSubworkoutParams{
						Out:          value,
						In:           r.PathValue("subworkout"),
						Superworkout: r.PathValue("superworkout"),
					}, nil
				},
			},
			"superworkout": &base.PatchReqParams[workoutdb.RawUpdateSubworkoutSuperworkoutParams]{
				Query: (*workoutdb.Queries).RawUpdateSubworkoutSuperworkout,
				Convert: func(r *http.Request, value string) (*workoutdb.RawUpdateSubworkoutSuperworkoutParams, error) {
					return &workoutdb.RawUpdateSubworkoutSuperworkoutParams{
						Out:        value,
						In:         r.PathValue("superworkout"),
						Subworkout: r.PathValue("subworkout"),
					}, nil
				},
			},
		},
	)
}

// HandlePostSubworkoutView godoc
//
//	@Summary		Create new subworkout relationship
//	@Description	Creates a new subworkout-superworkout relationship in the database
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			subworkout		formData	string	true	"Subworkout ID"
//	@Param			superworkout	formData	string	true	"Superworkout ID"
//	@Success		201				{string}	string	"HTML content"
//	@Failure		400				{string}	string	"Bad request"
//	@Failure		500				{string}	string	"Internal server error"
//	@Router			/view/data/subworkout [post]
func HandlePostSubworkoutView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePostDataTableView(
		state.ReadonlyDB,
		state.WriteDB,
		(*workoutdb.Queries).RawInsertSubworkout,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertSubworkoutParams, error) {
			return &workoutdb.RawInsertSubworkoutParams{
				Subworkout:   values.Get("subworkout"),
				Superworkout: values.Get("superworkout"),
			}, nil
		},
		func(ctx context.Context, q *workoutdb.Queries, item workoutdb.Subworkout) (*templates.DataTableRow, error) {
			workouts, err := q.ListAllIndividualWorkouts(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to list workouts: %w", err)
			}

			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/subworkout", item.Subworkout, item.Superworkout),
				DeleteEndpoint: util.UrlPathJoin("/view/data/subworkout", item.Subworkout, item.Superworkout),
				Values: []templates.DataTableValue{
					{Name: "subworkout", Type: templates.Select, Value: item.Subworkout, SelectOptions: workouts},
					{Name: "superworkout", Type: templates.Select, Value: item.Superworkout, SelectOptions: workouts},
				},
			}, nil
		},
	)
}

// HandleDeleteSubworkoutView godoc
//
//	@Summary		Delete subworkout relationship
//	@Description	Deletes a subworkout-superworkout relationship
//	@Tags			rawdata
//	@Param			subworkout		path		string	true	"Subworkout ID"
//	@Param			superworkout	path		string	true	"Superworkout ID"
//	@Success		200				{string}	string	"OK"
//	@Failure		400				{string}	string	"Bad request"
//	@Failure		500				{string}	string	"Internal server error"
//	@Router			/view/data/subworkout/{subworkout}/{superworkout} [delete]
func HandleDeleteSubworkoutView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleDeleteTableRowViewRequest(
		state.WriteDB,
		(*workoutdb.Queries).RawDeleteSubworkout,
		func(r *http.Request) (*workoutdb.RawDeleteSubworkoutParams, error) {
			return &workoutdb.RawDeleteSubworkoutParams{
				Subworkout:   r.PathValue("subworkout"),
				Superworkout: r.PathValue("superworkout"),
			}, nil
		},
	)
}
