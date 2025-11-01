package index

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/RyRose/uplog/internal/templates"
)

// HandleGetDataTabView godoc
//
//	@Summary		Get data tab view
//	@Description	Renders the data tab view with tabbed navigation for different data tables
//	@Tags			index
//	@Produce		html
//	@Param			tabX	path		string	false	"Tab X index"
//	@Param			tabY	path		string	false	"Tab Y index"
//	@Success		200		{string}	string	"HTML content"
//	@Failure		400		{string}	string	"Invalid tab index"
//	@Router			/view/tabs/data/{tabX}/{tabY} [get]
func HandleGetDataTabView() http.HandlerFunc {
	tabs := [][]templates.DataTab{
		{
			{Title: "Lifts", Endpoint: "/view/data/lift"},
			{Title: "Movements", Endpoint: "/view/data/movement"},
			{Title: "Muscles", Endpoint: "/view/data/muscle"},
			{Title: "Routines", Endpoint: "/view/data/routine"},
			{Title: "Workouts", Endpoint: "/view/data/workout"},
		},
		{
			{Title: "Variables", Endpoint: "/view/data/template_variable"},
			{Title: "Lift Groups", Endpoint: "/view/data/lift_group"},
		},
		{
			{Title: "Side Weight", Endpoint: "/view/data/side_weight"},
			{Title: "Progress", Endpoint: "/view/data/progress"},
			{Title: "Subworkouts", Endpoint: "/view/data/subworkout"},
		},
		{
			{Title: "Routine:Workout", Endpoint: "/view/data/routine_workout_mapping"},
			{Title: "Lift:Muscle", Endpoint: "/view/data/lift_muscle_mapping"},
			{Title: "Lift:Workout", Endpoint: "/view/data/lift_workout_mapping"},
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rawTabX, rawTabY := r.PathValue("tabX"), r.PathValue("tabY")
		if rawTabX == "" && rawTabY == "" {
			if err := templates.DataTabView(tabs, 0, 0).Render(ctx, w); err != nil {
				slog.WarnContext(ctx, "failed to render data tab view", "error", err)
			}
			return
		}
		tabX, err := strconv.Atoi(rawTabX)
		if err != nil || tabX < 0 || tabX >= len(tabs) {
			http.Error(w, "invalid tab index", http.StatusBadRequest)
			slog.ErrorContext(ctx, "invalid tab index", "tabX", r.PathValue("tabX"), "error", err)
			return
		}
		tabY, err := strconv.Atoi(rawTabY)
		if err != nil || tabY < 0 || tabY >= len(tabs[tabX]) {
			http.Error(w, "invalid tab index", http.StatusBadRequest)
			slog.ErrorContext(ctx, "invalid tab index", "tabY", r.PathValue("tabY"), "error", err)
			return
		}
		if err := templates.DataTabView(tabs, tabX, tabY).Render(ctx, w); err != nil {
			slog.WarnContext(ctx, "failed to render data tab view", "error", err)
		}
	}
}
