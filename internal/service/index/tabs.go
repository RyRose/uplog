package index

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/RyRose/uplog/internal/templates"
)

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
			templates.DataTabView(tabs, 0, 0).Render(ctx, w)
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
		templates.DataTabView(tabs, tabX, tabY).Render(ctx, w)
	}
}
