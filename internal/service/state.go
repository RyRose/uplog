package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/RyRose/uplog/internal/config"
	"github.com/RyRose/uplog/internal/sqlc"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus"
)

type State struct {
	ReadonlyDB, WriteDB *sql.DB
	PrometheusRegistry  *prometheus.Registry
	StartTimestamp      time.Time
}

func (s *State) Close() error {
	var errs []error
	if err := s.ReadonlyDB.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close readonly database: %w", err))
	}
	if err := s.WriteDB.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close write database: %w", err))
	}
	return errors.Join(errs...)
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

	slog.InfoContext(ctx, "applying migrations")
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

func NewState(ctx context.Context, cfg *config.Data) (*State, error) {
	start := time.Now()
	wDB, rDB, err := setupDatabases(ctx, cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to setup databases: %w", err)
	}
	return &State{
		ReadonlyDB:         rDB,
		WriteDB:            wDB,
		PrometheusRegistry: prometheus.NewRegistry(),
		StartTimestamp:     start,
	}, nil
}
