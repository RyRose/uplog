package service

import (
	"log/slog"
	"net/http"
	"path"

	"github.com/RyRose/uplog/internal/calendar"
	"github.com/RyRose/uplog/internal/templates"
)

func handleIndexPage(tab, cssQuery string, calendarSrv *calendar.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		code := r.URL.Query().Get("code")
		if code != "" {
			// FIXME: Include error in the html response.
			if err := calendarSrv.Init(ctx, code, r.URL.Query().Get("state")); err != nil {
				slog.WarnContext(ctx, "failed to initialize calendar service", "error", err)
			}
		}
		err := templates.IndexPage(
			cssQuery,
			path.Join("/view/tabs", tab, r.PathValue("tabX"), r.PathValue("tabY")),
		).Render(ctx, w)
		if err != nil {
			http.Error(w, "failed to write response", http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to write response", "error", err)
		}
	}
}
