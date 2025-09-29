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

func HandleGetTemplateVariableView(roDB *sql.DB) http.HandlerFunc {
	return base.HandleGetDataTableView(
		roDB,
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

func HandlePatchTemplateVariableView(wDB *sql.DB) http.HandlerFunc {
	return base.HandlePatchTableRowViewID(
		wDB,
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

func HandlePostTemplateVariableView(roDB, wDB *sql.DB) http.HandlerFunc {
	return base.HandlePostDataTableView(
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
