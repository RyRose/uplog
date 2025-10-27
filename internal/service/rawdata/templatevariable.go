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

// HandleGetTemplateVariableView godoc
//
//	@Summary		Get template variable data table view
//	@Description	Renders a paginated table view of template variables
//	@Tags			rawdata
//	@Produce		html
//	@Param			offset	query		integer	false	"Pagination offset"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/template_variable [get]
func HandleGetTemplateVariableView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleGetDataTableView(
		state.ReadonlyDB,
		base.TableViewMetadata{
			Headers: []string{"ID", "Value"},
			Post:    "/view/data/template_variable",
		},
		(*workoutdb.Queries).RawSelectTemplateVariablePage,
		func(limit, offset int64) workoutdb.RawSelectTemplateVariablePageParams {
			return workoutdb.RawSelectTemplateVariablePageParams{
				Limit:  limit,
				Offset: offset,
			}
		},
		func(_ context.Context, _ *sql.DB, items []workoutdb.TemplateVariable) ([]templates.DataTableRow, error) {
			var rows []templates.DataTableRow
			for _, item := range items {
				rows = append(rows, templates.DataTableRow{
					PatchEndpoint:  util.UrlPathJoin("/view/data/template_variable", item.ID),
					DeleteEndpoint: util.UrlPathJoin("/view/data/template_variable", item.ID),
					Values: []templates.DataTableValue{
						{Name: "id", Type: templates.InputString, Value: item.ID},
						{Name: "value", Type: templates.TextArea, Value: item.Value},
					},
				})
			}
			rows = append(rows, templates.DataTableRow{
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString},
					{Name: "value", Type: templates.TextArea},
				},
			})
			return rows, nil
		},
	)
}

// HandlePatchTemplateVariableView godoc
//
//	@Summary		Update template variable data
//	@Description	Updates specific fields of a template variable entry by ID
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Param			id		path		string	true	"Template variable ID"
//	@Param			id		formData	string	false	"New template variable ID"
//	@Param			value	formData	string	false	"Template variable value"
//	@Success		200		{string}	string	"OK"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/template_variable/{id} [patch]
func HandlePatchTemplateVariableView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePatchTableRowViewID(
		state.WriteDB,
		map[string]base.PatcherID{
			"id": &base.PatchIDParams[workoutdb.RawUpdateTemplateVariableIdParams]{
				Query: (*workoutdb.Queries).RawUpdateTemplateVariableId,
				Convert: func(id, value string) (*workoutdb.RawUpdateTemplateVariableIdParams, error) {
					return &workoutdb.RawUpdateTemplateVariableIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"value": &base.PatchIDParams[workoutdb.RawUpdateTemplateVariableValueParams]{
				Query: (*workoutdb.Queries).RawUpdateTemplateVariableValue,
				Convert: func(id, value string) (*workoutdb.RawUpdateTemplateVariableValueParams, error) {
					return &workoutdb.RawUpdateTemplateVariableValueParams{
						ID:    id,
						Value: value,
					}, nil
				},
			},
		},
	)
}

// HandlePostTemplateVariableView godoc
//
//	@Summary		Create new template variable
//	@Description	Creates a new template variable entry in the database
//	@Tags			rawdata
//	@Accept			x-www-form-urlencoded
//	@Produce		html
//	@Param			id		formData	string	true	"Template variable ID"
//	@Param			value	formData	string	true	"Template variable value"
//	@Success		201		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/view/data/template_variable [post]
func HandlePostTemplateVariableView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandlePostDataTableView(
		state.ReadonlyDB,
		state.WriteDB,
		(*workoutdb.Queries).RawInsertTemplateVariable,
		func(_ context.Context, values url.Values) (*workoutdb.RawInsertTemplateVariableParams, error) {
			return &workoutdb.RawInsertTemplateVariableParams{
				ID:    values.Get("id"),
				Value: values.Get("value"),
			}, nil
		},
		func(_ context.Context, _ *workoutdb.Queries, item workoutdb.TemplateVariable) (*templates.DataTableRow, error) {
			return &templates.DataTableRow{
				PatchEndpoint:  util.UrlPathJoin("/view/data/template_variable", item.ID),
				DeleteEndpoint: util.UrlPathJoin("/view/data/template_variable", item.ID),
				Values: []templates.DataTableValue{
					{Name: "id", Type: templates.InputString, Value: item.ID},
					{Name: "value", Type: templates.TextArea, Value: item.Value},
				},
			}, nil
		},
	)
}

// HandleDeleteTemplateVariableView godoc
//
//	@Summary		Delete template variable
//	@Description	Deletes a template variable entry by ID
//	@Tags			rawdata
//	@Param			id	path		string	true	"Template variable ID"
//	@Success		200	{string}	string	"OK"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/view/data/template_variable/{id} [delete]
func HandleDeleteTemplateVariableView(_ *config.Data, state *config.State) http.HandlerFunc {
	return base.HandleDeleteTableRowViewID(state.WriteDB, (*workoutdb.Queries).RawDeleteTemplateVariable)
}
