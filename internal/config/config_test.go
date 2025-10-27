package config

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set some environment variables for testing
	os.Setenv("DATABASE_PATH", "/test/db.db")
	os.Setenv("PORT", "9090")
	os.Setenv("DEBUG", "true")
	defer os.Unsetenv("DATABASE_PATH")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("DEBUG")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}

	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		log.Fatalf("os.Chdir(../..) error = %v", err)
	}
	cfg, err := Load(context.Background(), "./config/main.lua")
	if err := os.Chdir(origDir); err != nil {
		log.Fatalf("os.Chdir(%q) error = %v", origDir, err)
	}
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DatabasePath != "/test/db.db" {
		t.Errorf("DatabasePath = %v, want /test/db.db", cfg.DatabasePath)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %v, want 9090", cfg.Port)
	}

	if !cfg.Debug {
		t.Errorf("Debug = %v, want true", cfg.Debug)
	}
}

func TestLuaTypes(t *testing.T) {
	types := LuaTypes()

	if len(types) == 0 {
		t.Error("LuaTypes() returned empty slice")
	}

	// Check that Data is in the types
	foundData := false
	for _, typ := range types {
		if _, ok := typ.(Data); ok {
			foundData = true
			break
		}
	}

	if !foundData {
		t.Error("LuaTypes() did not include Data type")
	}
}
