package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	state *State) {

	ts := fmt.Sprint(time.Now().Unix())
	traceMux := &mux.Trace{Mux: rawMux}
	webMux := &mux.Web{Mux: traceMux}
	roDB := state.ReadonlyDB
	wDB := state.WriteDB

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
	// TODO: Use hash of css for cache busting instead of date.
	traceMux.HandleFunc("GET /{$}", index.HandleIndexPage("main", ts))
	traceMux.HandleFunc("GET /data/{$}", index.HandleIndexPage("data", ts))
	traceMux.HandleFunc("GET /data/{tabX}/{tabY}", index.HandleIndexPage("data", ts))

	// Main view.
	webMux.Handle("GET /view/tabs/main", index.HandleMainTab(roDB))
	webMux.Handle("GET /view/liftgroups", index.HandleGetLiftGroupListView(roDB))

	// Progress table
	webMux.Handle("GET /view/progresstable", index.HandleGetProgressTable(roDB))
	webMux.Handle("DELETE /view/progresstablerow/{id}", index.HandleDeleteProgress(wDB))
	webMux.Handle("POST /view/progresstablerow", index.HandleCreateProgress(wDB))

	// Routine table.
	webMux.Handle("GET /view/routinetable", index.HandleGetRoutineTable(roDB))

	// Progress form.
	webMux.Handle("GET /view/liftselect", index.HandleGetLiftSelect(roDB))
	webMux.Handle("GET /view/sideweightselect", index.HandleGetSideWeightSelect(roDB))
	webMux.Handle("GET /view/progressform", index.HandleGetProgressForm())
	webMux.Handle("POST /view/progressform", index.HandleCreateProgressForm(roDB))

	// Data view
	webMux.Handle("GET /view/tabs/data/{$}", index.HandleGetDataTabView())
	webMux.Handle("GET /view/tabs/data/{tabX}/{tabY}", index.HandleGetDataTabView())

	// Lift table view
	webMux.Handle("POST /view/data/lift", rawdata.HandlePostLiftView(roDB, wDB))
	webMux.Handle("PATCH /view/data/lift", rawdata.HandlePatchLiftView(wDB))
	webMux.Handle("PATCH /view/data/lift/{id}", rawdata.HandlePatchLiftView(wDB))
	webMux.Handle("DELETE /view/data/lift", rawdata.HandleDeleteLiftView(wDB))
	webMux.Handle("DELETE /view/data/lift/{id}", rawdata.HandleDeleteLiftView(wDB))
	webMux.Handle("GET /view/data/lift", rawdata.HandleGetLiftView(roDB))

	// Movement table view
	webMux.Handle("POST /view/data/movement", rawdata.HandlePostMovementView(roDB, wDB))
	webMux.Handle("PATCH /view/data/movement", rawdata.HandlePatchMovementView(wDB))
	webMux.Handle("PATCH /view/data/movement/{id}", rawdata.HandlePatchMovementView(wDB))
	webMux.Handle("DELETE /view/data/movement", rawdata.HandleDeleteMovementView(wDB))
	webMux.Handle("DELETE /view/data/movement/{id}", rawdata.HandleDeleteMovementView(wDB))
	webMux.Handle("GET /view/data/movement", rawdata.HandleGetMovementView(roDB))

	// Muscle table view
	webMux.Handle("POST /view/data/muscle", rawdata.HandlePostMuscleView(roDB, wDB))
	webMux.Handle("PATCH /view/data/muscle", rawdata.HandlePatchMuscleView(wDB))
	webMux.Handle("PATCH /view/data/muscle/{id}", rawdata.HandlePatchMuscleView(wDB))
	webMux.Handle("DELETE /view/data/muscle", rawdata.HandleDeleteMuscleView(wDB))
	webMux.Handle("DELETE /view/data/muscle/{id}", rawdata.HandleDeleteMuscleView(wDB))
	webMux.Handle("GET /view/data/muscle", rawdata.HandleGetMuscleView(roDB))

	// Routine table view
	webMux.Handle("POST /view/data/routine", rawdata.HandlePostRoutineView(roDB, wDB))
	webMux.Handle("PATCH /view/data/routine", rawdata.HandlePatchRoutineView(wDB))
	webMux.Handle("PATCH /view/data/routine/{id}", rawdata.HandlePatchRoutineView(wDB))
	webMux.Handle("DELETE /view/data/routine", rawdata.HandleDeleteRoutineView(wDB))
	webMux.Handle("DELETE /view/data/routine/{id}", rawdata.HandleDeleteRoutineView(wDB))
	webMux.Handle("GET /view/data/routine", rawdata.HandleGetRoutineView(roDB))

	// Side weight view
	webMux.Handle("POST /view/data/side_weight", rawdata.HandlePostSideWeightView(roDB, wDB))
	webMux.Handle("PATCH /view/data/side_weight", rawdata.HandlePatchSideWeightView(wDB))
	webMux.Handle("PATCH /view/data/side_weight/{id}", rawdata.HandlePatchSideWeightView(wDB))
	webMux.Handle("DELETE /view/data/side_weight", rawdata.HandleDeleteSideWeightView(wDB))
	webMux.Handle("DELETE /view/data/side_weight/{id}", rawdata.HandleDeleteSideWeightView(wDB))
	webMux.Handle("GET /view/data/side_weight", rawdata.HandleGetSideWeightView(roDB))

	// Template variable view
	webMux.Handle("POST /view/data/template_variable", rawdata.HandlePostTemplateVariableView(roDB, wDB))
	webMux.Handle("PATCH /view/data/template_variable", rawdata.HandlePatchTemplateVariableView(wDB))
	webMux.Handle("PATCH /view/data/template_variable/{id}", rawdata.HandlePatchTemplateVariableView(wDB))
	webMux.Handle("DELETE /view/data/template_variable", rawdata.HandleDeleteTemplateVariableView(wDB))
	webMux.Handle("DELETE /view/data/template_variable/{id}", rawdata.HandleDeleteTemplateVariableView(wDB))
	webMux.Handle("GET /view/data/template_variable", rawdata.HandleGetTemplateVariableView(roDB))

	// Workout view
	webMux.Handle("POST /view/data/workout", rawdata.HandlePostWorkoutView(roDB, wDB))
	webMux.Handle("PATCH /view/data/workout", rawdata.HandlePatchWorkoutView(wDB))
	webMux.Handle("PATCH /view/data/workout/{id}", rawdata.HandlePatchWorkoutView(wDB))
	webMux.Handle("DELETE /view/data/workout", rawdata.HandleDeleteWorkoutView(wDB))
	webMux.Handle("DELETE /view/data/workout/{id}", rawdata.HandleDeleteWorkoutView(wDB))
	webMux.Handle("GET /view/data/workout", rawdata.HandleGetWorkoutView(roDB))

	// Progress view
	webMux.Handle("POST /view/data/progress", rawdata.HandlePostProgressView(roDB, wDB))
	webMux.Handle("PATCH /view/data/progress", rawdata.HandlePatchProgressView(wDB))
	webMux.Handle("PATCH /view/data/progress/{id}", rawdata.HandlePatchProgressView(wDB))
	webMux.Handle("DELETE /view/data/progress/{id}", rawdata.HandleDeleteProgressView(wDB))
	webMux.Handle("GET /view/data/progress", rawdata.HandleGetProgressView(roDB))

	// Lift muscle mapping view
	webMux.Handle("POST /view/data/lift_muscle_mapping", rawdata.HandlePostLiftMuscleView(roDB, wDB))
	webMux.Handle("PATCH /view/data/lift_muscle_mapping/{lift}/{muscle}/{movement}", rawdata.HandlePatchLiftMuscleView(wDB))
	webMux.Handle("DELETE /view/data/lift_muscle_mapping/{lift}/{muscle}/{movement}", rawdata.HandleDeleteLiftMuscleView(wDB))
	webMux.Handle("GET /view/data/lift_muscle_mapping", rawdata.HandleGetLiftMuscleView(roDB))

	// Lift workout mapping view
	webMux.Handle("POST /view/data/lift_workout_mapping", rawdata.HandlePostLiftWorkoutView(roDB, wDB))
	webMux.Handle("PATCH /view/data/lift_workout_mapping/{lift}/{workout}", rawdata.HandlePatchLiftWorkoutView(wDB))
	webMux.Handle("DELETE /view/data/lift_workout_mapping/{lift}/{workout}", rawdata.HandleDeleteLiftWorkoutView(wDB))
	webMux.Handle("GET /view/data/lift_workout_mapping", rawdata.HandleGetLiftWorkoutView(roDB))

	// Routine workout mapping view
	webMux.Handle("POST /view/data/routine_workout_mapping", rawdata.HandlePostRoutineWorkoutView(roDB, wDB))
	webMux.Handle("PATCH /view/data/routine_workout_mapping/{routine}/{workout}", rawdata.HandlePatchRoutineWorkoutView(wDB))
	webMux.Handle("DELETE /view/data/routine_workout_mapping/{routine}/{workout}", rawdata.HandleDeleteRoutineWorkoutView(wDB))
	webMux.Handle("GET /view/data/routine_workout_mapping", rawdata.HandleGetRoutineWorkoutView(roDB))

	// Subworkout mapping view
	webMux.Handle("POST /view/data/subworkout", rawdata.HandlePostSubworkoutView(roDB, wDB))
	webMux.Handle("PATCH /view/data/subworkout/{subworkout}/{superworkout}", rawdata.HandlePatchSubworkoutView(wDB))
	webMux.Handle("DELETE /view/data/subworkout/{subworkout}/{superworkout}", rawdata.HandleDeleteSubworkoutView(wDB))
	webMux.Handle("GET /view/data/subworkout", rawdata.HandleGetSubworkoutView(roDB))

	// Lift group table view
	webMux.Handle("POST /view/data/lift_group", rawdata.HandlePostLiftGroupView(roDB, wDB))
	webMux.Handle("PATCH /view/data/lift_group", rawdata.HandlePatchLiftGroupView(wDB))
	webMux.Handle("PATCH /view/data/lift_group/{id}", rawdata.HandlePatchLiftGroupView(wDB))
	webMux.Handle("DELETE /view/data/lift_group", rawdata.HandleDeleteLiftGroupView(wDB))
	webMux.Handle("DELETE /view/data/lift_group/{id}", rawdata.HandleDeleteLiftGroupView(wDB))
	webMux.Handle("GET /view/data/lift_group", rawdata.HandleGetLiftGroupView(roDB))
}
