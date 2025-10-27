package service

import (
	"context"
	"net/http"

	"github.com/RyRose/uplog/internal/config"
	"github.com/RyRose/uplog/internal/service/index"
	"github.com/RyRose/uplog/internal/service/mux"
	"github.com/RyRose/uplog/internal/service/rawdata"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpswagger "github.com/swaggo/http-swagger/v2"
)

func AddRoutes(
	_ context.Context,
	rawMux *http.ServeMux,
	cfg *config.Data,
	state *config.State) {

	traceMux := &mux.Trace{Mux: rawMux}
	webMux := &mux.Web{Mux: traceMux}

	// Health check endpoint.
	rawMux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Prometheus metrics.
	traceMux.Handle("/metrics", promhttp.HandlerFor(state.PrometheusRegistry, promhttp.HandlerOpts{}))

	// Swagger docs.
	// TODO: Only enable swagger docs in "dev" mode.
	traceMux.Handle("GET /docs/", http.FileServer(http.Dir(".")))
	traceMux.Handle("GET /swagger/", httpswagger.Handler(httpswagger.URL(cfg.SwaggerURL)))

	// Vendor and non-vendor static assets.
	traceMux.Handle("GET /web/static/", http.FileServer(http.Dir(".")))
	traceMux.Handle("GET /web/vendor/", http.FileServer(http.Dir(".")))

	// Index pages.
	traceMux.HandleFunc("GET /{$}", index.HandleIndexPage("main", cfg, state))
	traceMux.HandleFunc("GET /data/{$}", index.HandleIndexPage("data", cfg, state))
	traceMux.HandleFunc("GET /data/{tabX}/{tabY}", index.HandleIndexPage("data", cfg, state))

	// Main view.
	webMux.Handle("GET /view/tabs/main", index.HandleMainTab(cfg, state))
	webMux.Handle("GET /view/liftgroups", index.HandleGetLiftGroupListView(cfg, state))

	// Progress table
	webMux.Handle("GET /view/progresstable", index.HandleGetProgressTable(cfg, state))
	webMux.Handle("DELETE /view/progresstablerow/{id}", index.HandleDeleteProgress(cfg, state))
	webMux.Handle("POST /view/progresstablerow", index.HandleCreateProgress(cfg, state))

	// Routine table.
	webMux.Handle("GET /view/routinetable", index.HandleGetRoutineTable(cfg, state))

	// Progress form.
	webMux.Handle("GET /view/liftselect", index.HandleGetLiftSelect(cfg, state))
	webMux.Handle("GET /view/sideweightselect", index.HandleGetSideWeightSelect(cfg, state))
	webMux.Handle("GET /view/progressform", index.HandleGetProgressForm())
	webMux.Handle("POST /view/progressform", index.HandleCreateProgressForm(cfg, state))

	// Data view
	webMux.Handle("GET /view/tabs/data/{$}", index.HandleGetDataTabView())
	webMux.Handle("GET /view/tabs/data/{tabX}/{tabY}", index.HandleGetDataTabView())

	// Lift table view
	webMux.Handle("POST /view/data/lift", rawdata.HandlePostLiftView(cfg, state))
	webMux.Handle("PATCH /view/data/lift", rawdata.HandlePatchLiftView(cfg, state))
	webMux.Handle("PATCH /view/data/lift/{id}", rawdata.HandlePatchLiftView(cfg, state))
	webMux.Handle("DELETE /view/data/lift", rawdata.HandleDeleteLiftView(cfg, state))
	webMux.Handle("DELETE /view/data/lift/{id}", rawdata.HandleDeleteLiftView(cfg, state))
	webMux.Handle("GET /view/data/lift", rawdata.HandleGetLiftView(cfg, state))

	// Movement table view
	webMux.Handle("POST /view/data/movement", rawdata.HandlePostMovementView(cfg, state))
	webMux.Handle("PATCH /view/data/movement", rawdata.HandlePatchMovementView(cfg, state))
	webMux.Handle("PATCH /view/data/movement/{id}", rawdata.HandlePatchMovementView(cfg, state))
	webMux.Handle("DELETE /view/data/movement", rawdata.HandleDeleteMovementView(cfg, state))
	webMux.Handle("DELETE /view/data/movement/{id}", rawdata.HandleDeleteMovementView(cfg, state))
	webMux.Handle("GET /view/data/movement", rawdata.HandleGetMovementView(cfg, state))

	// Muscle table view
	webMux.Handle("POST /view/data/muscle", rawdata.HandlePostMuscleView(cfg, state))
	webMux.Handle("PATCH /view/data/muscle", rawdata.HandlePatchMuscleView(cfg, state))
	webMux.Handle("PATCH /view/data/muscle/{id}", rawdata.HandlePatchMuscleView(cfg, state))
	webMux.Handle("DELETE /view/data/muscle", rawdata.HandleDeleteMuscleView(cfg, state))
	webMux.Handle("DELETE /view/data/muscle/{id}", rawdata.HandleDeleteMuscleView(cfg, state))
	webMux.Handle("GET /view/data/muscle", rawdata.HandleGetMuscleView(cfg, state))

	// Routine table view
	webMux.Handle("POST /view/data/routine", rawdata.HandlePostRoutineView(cfg, state))
	webMux.Handle("PATCH /view/data/routine", rawdata.HandlePatchRoutineView(cfg, state))
	webMux.Handle("PATCH /view/data/routine/{id}", rawdata.HandlePatchRoutineView(cfg, state))
	webMux.Handle("DELETE /view/data/routine", rawdata.HandleDeleteRoutineView(cfg, state))
	webMux.Handle("DELETE /view/data/routine/{id}", rawdata.HandleDeleteRoutineView(cfg, state))
	webMux.Handle("GET /view/data/routine", rawdata.HandleGetRoutineView(cfg, state))

	// Side weight view
	webMux.Handle("POST /view/data/side_weight", rawdata.HandlePostSideWeightView(cfg, state))
	webMux.Handle("PATCH /view/data/side_weight", rawdata.HandlePatchSideWeightView(cfg, state))
	webMux.Handle("PATCH /view/data/side_weight/{id}", rawdata.HandlePatchSideWeightView(cfg, state))
	webMux.Handle("DELETE /view/data/side_weight", rawdata.HandleDeleteSideWeightView(cfg, state))
	webMux.Handle("DELETE /view/data/side_weight/{id}", rawdata.HandleDeleteSideWeightView(cfg, state))
	webMux.Handle("GET /view/data/side_weight", rawdata.HandleGetSideWeightView(cfg, state))

	// Template variable view
	webMux.Handle("POST /view/data/template_variable", rawdata.HandlePostTemplateVariableView(cfg, state))
	webMux.Handle("PATCH /view/data/template_variable", rawdata.HandlePatchTemplateVariableView(cfg, state))
	webMux.Handle("PATCH /view/data/template_variable/{id}", rawdata.HandlePatchTemplateVariableView(cfg, state))
	webMux.Handle("DELETE /view/data/template_variable", rawdata.HandleDeleteTemplateVariableView(cfg, state))
	webMux.Handle("DELETE /view/data/template_variable/{id}", rawdata.HandleDeleteTemplateVariableView(cfg, state))
	webMux.Handle("GET /view/data/template_variable", rawdata.HandleGetTemplateVariableView(cfg, state))

	// Workout view
	webMux.Handle("POST /view/data/workout", rawdata.HandlePostWorkoutView(cfg, state))
	webMux.Handle("PATCH /view/data/workout", rawdata.HandlePatchWorkoutView(cfg, state))
	webMux.Handle("PATCH /view/data/workout/{id}", rawdata.HandlePatchWorkoutView(cfg, state))
	webMux.Handle("DELETE /view/data/workout", rawdata.HandleDeleteWorkoutView(cfg, state))
	webMux.Handle("DELETE /view/data/workout/{id}", rawdata.HandleDeleteWorkoutView(cfg, state))
	webMux.Handle("GET /view/data/workout", rawdata.HandleGetWorkoutView(cfg, state))

	// Progress view
	webMux.Handle("POST /view/data/progress", rawdata.HandlePostProgressView(cfg, state))
	webMux.Handle("PATCH /view/data/progress", rawdata.HandlePatchProgressView(cfg, state))
	webMux.Handle("PATCH /view/data/progress/{id}", rawdata.HandlePatchProgressView(cfg, state))
	webMux.Handle("DELETE /view/data/progress/{id}", rawdata.HandleDeleteProgressView(cfg, state))
	webMux.Handle("GET /view/data/progress", rawdata.HandleGetProgressView(cfg, state))

	// Lift muscle mapping view
	webMux.Handle("POST /view/data/lift_muscle_mapping", rawdata.HandlePostLiftMuscleView(cfg, state))
	webMux.Handle("PATCH /view/data/lift_muscle_mapping/{lift}/{muscle}/{movement}", rawdata.HandlePatchLiftMuscleView(cfg, state))
	webMux.Handle("DELETE /view/data/lift_muscle_mapping/{lift}/{muscle}/{movement}", rawdata.HandleDeleteLiftMuscleView(cfg, state))
	webMux.Handle("GET /view/data/lift_muscle_mapping", rawdata.HandleGetLiftMuscleView(cfg, state))

	// Lift workout mapping view
	webMux.Handle("POST /view/data/lift_workout_mapping", rawdata.HandlePostLiftWorkoutView(cfg, state))
	webMux.Handle("PATCH /view/data/lift_workout_mapping/{lift}/{workout}", rawdata.HandlePatchLiftWorkoutView(cfg, state))
	webMux.Handle("DELETE /view/data/lift_workout_mapping/{lift}/{workout}", rawdata.HandleDeleteLiftWorkoutView(cfg, state))
	webMux.Handle("GET /view/data/lift_workout_mapping", rawdata.HandleGetLiftWorkoutView(cfg, state))

	// Routine workout mapping view
	webMux.Handle("POST /view/data/routine_workout_mapping", rawdata.HandlePostRoutineWorkoutView(cfg, state))
	webMux.Handle("PATCH /view/data/routine_workout_mapping/{routine}/{workout}", rawdata.HandlePatchRoutineWorkoutView(cfg, state))
	webMux.Handle("DELETE /view/data/routine_workout_mapping/{routine}/{workout}", rawdata.HandleDeleteRoutineWorkoutView(cfg, state))
	webMux.Handle("GET /view/data/routine_workout_mapping", rawdata.HandleGetRoutineWorkoutView(cfg, state))

	// Subworkout mapping view
	webMux.Handle("POST /view/data/subworkout", rawdata.HandlePostSubworkoutView(cfg, state))
	webMux.Handle("PATCH /view/data/subworkout/{subworkout}/{superworkout}", rawdata.HandlePatchSubworkoutView(cfg, state))
	webMux.Handle("DELETE /view/data/subworkout/{subworkout}/{superworkout}", rawdata.HandleDeleteSubworkoutView(cfg, state))
	webMux.Handle("GET /view/data/subworkout", rawdata.HandleGetSubworkoutView(cfg, state))

	// Lift group table view
	webMux.Handle("POST /view/data/lift_group", rawdata.HandlePostLiftGroupView(cfg, state))
	webMux.Handle("PATCH /view/data/lift_group", rawdata.HandlePatchLiftGroupView(cfg, state))
	webMux.Handle("PATCH /view/data/lift_group/{id}", rawdata.HandlePatchLiftGroupView(cfg, state))
	webMux.Handle("DELETE /view/data/lift_group", rawdata.HandleDeleteLiftGroupView(cfg, state))
	webMux.Handle("DELETE /view/data/lift_group/{id}", rawdata.HandleDeleteLiftGroupView(cfg, state))
	webMux.Handle("GET /view/data/lift_group", rawdata.HandleGetLiftGroupView(cfg, state))
}
