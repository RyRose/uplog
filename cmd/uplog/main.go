package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/RyRose/uplog/internal/calendar"
	"github.com/RyRose/uplog/internal/service"
	"github.com/RyRose/uplog/internal/sqlc"
	"github.com/RyRose/uplog/internal/sqlc/workoutdb"
	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

func todaysDate() time.Time {
	return time.Now()
}

func setupDatabases(ctx context.Context, dbPath string) (*sql.DB, *sql.DB, error) {

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directories: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?mode=rwc&_journal_mode=WAL&_txlock=immediate", url.QueryEscape(dbPath))
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open %v: %w", dsn, err)
	}
	db.SetMaxOpenConns(1)

	goose.SetBaseFS(sqlc.EmbedMigrations)
	if err := goose.SetDialect("sqlite"); err != nil {
		return nil, nil, fmt.Errorf("failed to set dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return nil, nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	rdsn := fmt.Sprintf("file:%s?mode=ro&_journal_mode=WAL&_txlock=immediate", url.QueryEscape(dbPath))
	rdb, err := sql.Open("sqlite3", rdsn)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to open %v: %w", rdsn, err)
	}
	return db, rdb, nil
}

func initializeCalendarEvents(ctx context.Context, srv *calendar.Service, db *sql.DB) {
	if !srv.Initialized() {
		slog.WarnContext(ctx, "calendar service not initialized")
		return
	}

	queries := workoutdb.New(db)
	internalEvents, err := queries.ListCurrentSchedule(ctx, todaysDate().Format(time.DateOnly))
	if err != nil {
		slog.WarnContext(ctx, "failed to list current schedule", "error", err)
	}
	var date string
	for i, evt := range internalEvents {
		if evt.Workout == nil {
			continue
		}
		date = evt.Date
		slog.DebugContext(ctx, "syncing event", "event", evt)
		if err := srv.Sync(ctx, calendar.Event{
			Summary:     evt.Workout.(string),
			ISO8601Date: evt.Date,
			Description: os.Getenv("CALENDAR_DESCRIPTION"),
		}); err != nil {
			slog.WarnContext(ctx, "failed to sync event", "event", evt, "index", i, "error", err)
		}
	}
	if date == "" {
		return
	}
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse date", "date", date, "error", err)
		return
	}
	if err := srv.DeleteAfter(ctx, t.Add(48*time.Hour)); err != nil {
		slog.WarnContext(ctx, "failed to delete events", "date", t, "error", err)
	}
}

func envOrDefault(key, def string) string {
	value := os.Getenv(key)
	if value == "" {
		value = def
	}
	return value
}

func makeCalendarService(ctx context.Context, oauthCredentials string, oauthTokenPath string) (*calendar.Service, error) {
	if oauthCredentials == "" {
		return nil, errors.New("oauth credentials not provided")
	}
	calendarService, err := calendar.NewService(ctx, []byte(oauthCredentials), oauthTokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}
	return calendarService, nil
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(
		ctx, os.Interrupt,
		syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer cancel()

	if os.Getenv("DEBUG") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	if os.Getenv("VERSION") != "" {
		slog.InfoContext(ctx, "uplog version", "version", os.Getenv("VERSION"))
	} else {
		slog.WarnContext(ctx, "no uplog version set")
	}

	dbPath := envOrDefault("DATABASE_PATH", "./tmp/db/data.db")
	db, rdb, err := setupDatabases(ctx, dbPath)
	if err != nil {
		return fmt.Errorf("failed to setup databases: %w", err)
	}
	defer db.Close()
	defer rdb.Close()

	oauthTokenPath := envOrDefault("OAUTH_TOKEN_PATH", "./secrets/oauth_token.json")

	calendarService, err := makeCalendarService(
		ctx,
		os.Getenv("OAUTH_CREDENTIALS"),
		oauthTokenPath)
	if err != nil {
		slog.WarnContext(ctx, "failed to create calendar service", "error", err)
		calendarService = nil
	}
	defer calendarService.Close()
	go initializeCalendarEvents(ctx, calendarService, db)

	srv := &http.Server{
		Addr:    net.JoinHostPort("", envOrDefault("PORT", "8080")),
		Handler: service.NewServer(ctx, db, rdb, calendarService),
	}
	go func(ctx context.Context) {
		<-ctx.Done()
		slog.InfoContext(ctx, "server context done, shutting down server")
		if err := srv.Shutdown(ctx); err != nil {
			slog.WarnContext(ctx, "failed to shutdown server", "error", err)
		}
	}(ctx)

	slog.InfoContext(ctx, "start listening",
		"addr", srv.Addr,
		"oauth_token_path", oauthTokenPath,
		"db_path", dbPath,
		"debug_log", slog.Default().Enabled(ctx, slog.LevelDebug))
	err = srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("failed to listen and serve: %w", srv.ListenAndServe())
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.ErrorContext(ctx, "server failed to run", "error", err)
		os.Exit(1)
	}
	slog.InfoContext(ctx, "server exited gracefully")
}
