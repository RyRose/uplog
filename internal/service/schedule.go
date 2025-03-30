package service

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/RyRose/uplog/internal/calendar"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/RyRose/uplog/internal/templates"
)

func scheduleView(s workoutdb.Schedule, workoutOptions []string) templates.ScheduleDataView {
	var weekday string
	d, err := time.Parse(time.DateOnly, s.Date)
	if err != nil {
		slog.Warn("failed to parse date", "schedule", s, "error", err)
	} else {
		weekday = d.Weekday().String()[:3]
	}
	return templates.ScheduleDataView{
		Date:    s.Date,
		Workout: s.Workout.(string),
		Weekday: weekday,
		Options: workoutOptions,
	}
}

func scheduleViews(s []workoutdb.Schedule, workoutOptions []string) []templates.ScheduleDataView {
	var out []templates.ScheduleDataView
	for _, schedule := range s {
		out = append(out, scheduleView(schedule, workoutOptions))
	}
	return out
}

func handleGetScheduleTable(roDB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(roDB)
		schedule, err := queries.ListCurrentSchedule(ctx, todaysDate().Format(time.DateOnly))
		if err != nil {
			slog.ErrorContext(ctx, "failed to list current schedule", "error", err)
			http.Error(w, "failed to list current schedule", http.StatusInternalServerError)
			return
		}

		workouts, err := queries.ListAllIndividualWorkouts(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to list all individual workouts", "error", err)
			http.Error(w, "failed to list all individual workouts", http.StatusInternalServerError)
			return
		}

		workoutLists, err := queries.ListAllIndividualScheduleLists(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to list all individual schedule lists", "error", err)
			http.Error(w, "failed to list all individual schedule lists", http.StatusInternalServerError)
			return
		}

		templates.ScheduleTable(scheduleViews(schedule, workouts), workouts, workoutLists).Render(ctx, w)
	}
}

func handlePostScheduleTableRows(wDB *sql.DB, calendarService *calendar.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(wDB)

		if err := r.ParseForm(); err != nil {
			slog.ErrorContext(ctx, "failed to parse form", "error", err)
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		tx, err := wDB.BeginTx(ctx, nil)
		if err != nil {
			slog.ErrorContext(ctx, "failed to begin transaction", "error", err)
			http.Error(w, "failed to begin transaction", http.StatusInternalServerError)
			return
		}

		var dates []string
		for _, date := range r.PostForm["date"] {
			dates = append(dates, strings.TrimSpace(date))
		}
		var workouts []string
		dateMapping := make(map[string]string)
		for i, workout := range r.PostForm["workout"] {
			workouts = append(workouts, strings.TrimSpace(workout))
			dateMapping[dates[i]] = workout
		}

		if len(dates) != len(workouts) {
			slog.ErrorContext(ctx, "mismatched dates and workouts", "dates", len(dates), "workouts", len(workouts))
			http.Error(w, "mismatched dates and workouts", http.StatusBadRequest)
			return
		}

		var minimumDateTime time.Time
		for _, date := range dates {
			t, err := time.Parse(time.DateOnly, date)
			if err != nil {
				slog.ErrorContext(ctx, "failed to parse date", "date", date, "error", err)
				http.Error(w, "failed to parse date", http.StatusBadRequest)
				return
			}
			if minimumDateTime.IsZero() || t.Before(minimumDateTime) {
				minimumDateTime = t
			}
		}

		qTx := queries.WithTx(tx)

		for i, date := range dates {
			newDate := minimumDateTime.Add(time.Duration(i*24) * time.Hour).Format(time.DateOnly)
			if date == newDate || dateMapping[newDate] == workouts[i] {
				continue
			}
			qTx.RawUpdateScheduleWorkout(ctx, workoutdb.RawUpdateScheduleWorkoutParams{
				Date:    newDate,
				Workout: workouts[i],
			})
			if err := calendarService.Sync(ctx, calendar.Event{
				Summary:     workouts[i],
				ISO8601Date: newDate,
				Description: "https://workout.ryanrose.me",
			}); err != nil {
				slog.WarnContext(ctx, "failed to sync event", "workout", workouts[i], "date", date, "newDate", newDate, "error", err)
			}
		}

		if err := tx.Commit(); err != nil {
			slog.ErrorContext(ctx, "failed to commit transaction", "error", err)
			http.Error(w, "failed to commit transaction", http.StatusInternalServerError)
			return
		}

		schedule, err := queries.ListCurrentSchedule(ctx, todaysDate().Format(time.DateOnly))
		if err != nil {
			slog.ErrorContext(ctx, "failed to list current schedule", "error", err)
			http.Error(w, "failed to list current schedule", http.StatusInternalServerError)
			return
		}
		allWorkouts, err := queries.ListAllIndividualWorkouts(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to list all individual workouts", "error", err)
			http.Error(w, "failed to list all individual workouts", http.StatusInternalServerError)
			return
		}
		templates.ScheduleTableRows(scheduleViews(schedule, allWorkouts)).Render(ctx, w)
	}
}

func handleDeleteSchedule(wDB *sql.DB, calendarService *calendar.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(wDB)

		dateParam := r.PathValue("date")

		// TODO: Wrap queries in a transaction.

		if err := queries.DeleteSchedule(ctx, dateParam); err != nil {
			slog.ErrorContext(ctx, "failed to delete schedule", "error", err)
			http.Error(w, "failed to delete schedule", http.StatusInternalServerError)
		}

		schedule, err := queries.ListCurrentSchedule(ctx, dateParam)
		if err != nil {
			slog.ErrorContext(ctx, "failed to list schedule after date", "error", err, "date", dateParam)
			http.Error(w, "failed to list current schedule", http.StatusInternalServerError)
			return
		}

		lastScheduleDate := dateParam
		for _, s := range schedule {
			lastScheduleDate = s.Date
			date, _ := time.Parse(time.DateOnly, s.Date)
			toDate := date.Add(-24 * time.Hour).Format(time.DateOnly)
			_, err := queries.UpdateScheduleDate(ctx, workoutdb.UpdateScheduleDateParams{
				FromDate: s.Date,
				ToDate:   toDate,
			})
			if err != nil {
				slog.ErrorContext(ctx, "failed to update schedule date", "error", err)
				http.Error(w, "failed to update schedule date", http.StatusInternalServerError)
				return
			}
			if err := calendarService.Sync(ctx, calendar.Event{
				Summary:     s.Workout.(string),
				ISO8601Date: toDate,
				Description: "https://workout.ryanrose.me",
			}); err != nil {
				slog.WarnContext(ctx, "failed to sync event", "event", s, "error", err)
			}
		}
		if err := calendarService.Sync(ctx, calendar.Event{ISO8601Date: lastScheduleDate}); err != nil {
			slog.WarnContext(ctx, "failed to sync event", "date", dateParam, "error", err)
		}

		schedule, err = queries.ListCurrentSchedule(ctx, todaysDate().Format(time.DateOnly))
		if err != nil {
			slog.ErrorContext(ctx, "failed to list current schedule", "error", err)
			http.Error(w, "failed to list current schedule", http.StatusInternalServerError)
			return
		}
		workouts, err := queries.ListAllIndividualWorkouts(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to list all individual workouts", "error", err)
			http.Error(w, "failed to list all individual workouts", http.StatusInternalServerError)
			return
		}
		templates.ScheduleTableOnly(scheduleViews(schedule, workouts)).Render(ctx, w)
	}
}

func handlePatchScheduleTableRow(wDB *sql.DB, calendarService *calendar.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(wDB)

		params := workoutdb.UpdateScheduleParams{
			Date:    r.PathValue("date"),
			Workout: r.PostFormValue("workout"),
			Notes:   r.PostFormValue("notes"),
		}

		schedule, err := queries.UpdateSchedule(ctx, params)
		if err != nil {
			slog.ErrorContext(ctx, "failed to update schedule", "error", err, "params", params)
			http.Error(w, "failed to update schedule", http.StatusBadRequest)
			return
		}

		if err := calendarService.Sync(ctx, calendar.Event{
			Summary:     schedule.Workout.(string),
			ISO8601Date: schedule.Date,
			Description: "https://workout.ryanrose.me",
		}); err != nil {
			slog.WarnContext(ctx, "failed to sync event", "event", schedule, "error", err)
		}

		workouts, err := queries.ListAllIndividualWorkouts(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to list all individual workouts", "error", err)
			http.Error(w, "failed to list all individual workouts", http.StatusInternalServerError)
			return
		}
		templates.ScheduleTableRow(scheduleView(schedule, workouts)).Render(ctx, w)
	}
}

func handlePostScheduleTableRow(wDB *sql.DB, calendarService *calendar.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := workoutdb.New(wDB)

		var date time.Time
		rawDate, err := queries.GetLatestScheduleDate(ctx)
		switch err {
		case sql.ErrNoRows:
			date = todaysDate()
		case nil:
			date, err = time.Parse(time.DateOnly, rawDate)
			if err != nil {
				http.Error(w, "failed to parse date", http.StatusBadRequest)
				slog.ErrorContext(ctx, "failed to parse date", "date", rawDate, "error", err)
				return
			}
			date = date.Add(24 * time.Hour)
			if date.Before(todaysDate()) {
				date = todaysDate()
			}
		default:
			http.Error(w, "failed to get latest schedule date", http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to get latest schedule date", "error", err)
			return
		}

		allWorkouts, err := queries.ListAllIndividualWorkouts(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to list all individual workouts", "error", err)
			http.Error(w, "failed to list all individual workouts", http.StatusInternalServerError)
			return
		}

		workoutList := r.PostFormValue("workout_list")
		if workoutList == "" {
			params := workoutdb.InsertScheduleParams{
				Date:    date.Format(time.DateOnly),
				Workout: r.PostFormValue("workout"),
			}

			schedule, err := queries.InsertSchedule(ctx, params)
			if err != nil {
				slog.ErrorContext(ctx, "failed to insert schedule", "error", err, "params", params)
				http.Error(w, "failed to insert schedule", http.StatusBadRequest)
				return
			}

			if err := calendarService.Sync(ctx, calendar.Event{
				Summary:     schedule.Workout.(string),
				ISO8601Date: schedule.Date,
				Description: "https://workout.ryanrose.me",
			}); err != nil {
				slog.WarnContext(ctx, "failed to sync event", "event", schedule, "error", err)
			}

			templates.ScheduleTableRow(scheduleView(schedule, allWorkouts)).Render(ctx, w)
			return
		}

		workouts, err := queries.ListWorkoutsForScheduleList(ctx, workoutList)
		if err != nil {
			http.Error(w, "failed to list workouts for schedule list", http.StatusBadRequest)
			slog.ErrorContext(ctx, "failed to list workouts for schedule list", "error", err)
			return
		}

		for _, workout := range workouts {
			params := workoutdb.InsertScheduleParams{
				Date:    date.Format(time.DateOnly),
				Workout: workout,
			}
			schedule, err := queries.InsertSchedule(ctx, params)
			if err != nil {
				slog.ErrorContext(ctx, "failed to insert schedule", "error", err, "params", params)
				http.Error(w, "failed to insert schedule", http.StatusBadRequest)
				return
			}

			event := calendar.Event{
				Summary:     schedule.Workout.(string),
				ISO8601Date: schedule.Date,
				Description: "https://workout.ryanrose.me",
			}
			if err := calendarService.Sync(ctx, event); err != nil {
				slog.WarnContext(ctx, "failed to sync events", "event", event, "error", err)
			}
			templates.ScheduleTableRow(scheduleView(schedule, allWorkouts)).Render(ctx, w)
			date = date.Add(24 * time.Hour)
		}
	}
}

func handleGetCalendarAuthURL(calendarService *calendar.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !calendarService.Initializable() || calendarService.Initialized() {
			w.WriteHeader(http.StatusOK)
			return
		}

		ctx := r.Context()
		url, err := calendarService.AuthCodeURL()
		if err != nil {
			http.Error(w, "failed to get auth code url", http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to get auth code url", "error", err)
			return
		}
		if err := templates.AuthorizationURL(url).Render(ctx, w); err != nil {
			http.Error(w, "failed to write response", http.StatusInternalServerError)
			slog.ErrorContext(ctx, "failed to write response", "error", err)
		}
	}
}
