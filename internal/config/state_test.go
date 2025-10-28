package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewState(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &Data{
		DatabasePath: dbPath,
	}

	state, err := NewState(ctx, cfg)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}
	defer state.Close()

	if state.RDB == nil {
		t.Error("ReadonlyDB is nil")
	}
	if state.WDB == nil {
		t.Error("WriteDB is nil")
	}
	if state.PrometheusRegistry == nil {
		t.Error("PrometheusRegistry is nil")
	}

	// Verify readonly DB is actually readonly
	_, err = state.RDB.Exec("CREATE TABLE test (id INTEGER)")
	if err == nil {
		t.Error("expected readonly DB to reject writes, but it didn't")
	}

	// Verify write DB can write
	_, err = state.WDB.Exec("CREATE TABLE test (id INTEGER)")
	if err != nil {
		t.Errorf("write DB should allow writes, got error: %v", err)
	}
}

func TestNewState_InvalidPath(t *testing.T) {
	ctx := context.Background()
	cfg := &Data{
		DatabasePath: "/invalid/path/that/cannot/exist/\x00/test.db",
	}

	_, err := NewState(ctx, cfg)
	if err == nil {
		t.Error("NewState() should fail with invalid path, but succeeded")
	}
}

func TestNewState_CreatesDirectories(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir1", "subdir2", "test.db")

	cfg := &Data{
		DatabasePath: dbPath,
	}

	state, err := NewState(ctx, cfg)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}
	defer state.Close()

	// Verify the directory was created
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("NewState() did not create parent directories")
	}
}

func TestState_Close(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &Data{
		DatabasePath: dbPath,
	}

	state, err := NewState(ctx, cfg)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}

	err = state.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify DBs are closed by trying to ping them
	err = state.RDB.Ping()
	if err == nil {
		t.Error("ReadonlyDB should be closed, but Ping succeeded")
	}

	err = state.WDB.Ping()
	if err == nil {
		t.Error("WriteDB should be closed, but Ping succeeded")
	}
}

func TestState_Close_Multiple(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &Data{
		DatabasePath: dbPath,
	}

	state, err := NewState(ctx, cfg)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}

	// Close once
	err = state.Close()
	if err != nil {
		t.Errorf("first Close() error = %v", err)
	}

	// Close again should not panic (sql.DB.Close is idempotent)
	err = state.Close()
	// Multiple closes may or may not return an error, but shouldn't panic
	if err != nil {
		t.Logf("second Close() returned error (expected): %v", err)
	}
}

func TestSetupDatabases(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	wDB, rDB, err := setupDatabases(ctx, dbPath)
	if err != nil {
		t.Fatalf("setupDatabases() error = %v", err)
	}
	defer wDB.Close()
	defer rDB.Close()

	if wDB == nil {
		t.Error("write DB is nil")
	}
	if rDB == nil {
		t.Error("readonly DB is nil")
	}

	// Verify WAL mode is enabled on write DB
	var journalMode string
	err = wDB.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Errorf("failed to query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("journal_mode = %v, want wal", journalMode)
	}
}

func TestSetupDatabases_MigrationsApplied(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	wDB, rDB, err := setupDatabases(ctx, dbPath)
	if err != nil {
		t.Fatalf("setupDatabases() error = %v", err)
	}
	defer wDB.Close()
	defer rDB.Close()

	// Verify goose_db_version table exists (created by goose migrations)
	var tableName string
	err = wDB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='goose_db_version'").Scan(&tableName)
	if err != nil {
		t.Errorf("goose_db_version table not found: %v", err)
	}
}

func TestSetupDatabases_WriteDBMaxConnections(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	wDB, rDB, err := setupDatabases(ctx, dbPath)
	if err != nil {
		t.Fatalf("setupDatabases() error = %v", err)
	}
	defer wDB.Close()
	defer rDB.Close()

	// The max open conns should be set to 1
	stats := wDB.Stats()
	if stats.MaxOpenConnections != 1 {
		t.Errorf("MaxOpenConnections = %d, want 1", stats.MaxOpenConnections)
	}
}
