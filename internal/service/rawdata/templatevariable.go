package rawdata

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"

	"github.com/RyRose/uplog/internal/service/rawdata/util"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func HandleGetTemplateVariableView(roDB *sql.DB) http.HandlerFunc {
	return handleGetDataTableView(
		roDB,
		tableViewMetadata{
			headers: []string{"ID", "Value"},
			post:    "/view/data/template_variable",
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

func HandlePatchTemplateVariableView(wDB *sql.DB) http.HandlerFunc {
	return handlePatchTableRowViewID(
		wDB,
		map[string]patcherID{
			"id": &patchIDParams[workoutdb.RawUpdateTemplateVariableIdParams]{
				query: (*workoutdb.Queries).RawUpdateTemplateVariableId,
				convert: func(id, value string) (*workoutdb.RawUpdateTemplateVariableIdParams, error) {
					return &workoutdb.RawUpdateTemplateVariableIdParams{
						In:  id,
						Out: value,
					}, nil
				},
			},
			"value": &patchIDParams[workoutdb.RawUpdateTemplateVariableValueParams]{
				query: (*workoutdb.Queries).RawUpdateTemplateVariableValue,
				convert: func(id, value string) (*workoutdb.RawUpdateTemplateVariableValueParams, error) {
					return &workoutdb.RawUpdateTemplateVariableValueParams{
						ID:    id,
						Value: value,
					}, nil
				},
			},
		},
	)
}

func HandlePostTemplateVariableView(roDB, wDB *sql.DB) http.HandlerFunc {
	return handlePostDataTableView(
		roDB,
		wDB,
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
