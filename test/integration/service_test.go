package integration

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/RyRose/uplog/internal/service"
	_ "github.com/mattn/go-sqlite3"
)

func TestIntegration_ServiceStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	port, err := getFreePort()
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}

	os.Setenv("PORT", fmt.Sprintf("%d", port))
	os.Setenv("DATABASE_PATH", dbPath)
	os.Setenv("DEBUG", "false")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_PATH")
		os.Unsetenv("DEBUG")
	}()

	// Change to repo root so Lua's require() can find config modules
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("failed to get repo root: %v", err)
	}

	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change to repo root: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		if err := service.Run(ctx, "./config/main.lua"); err != nil {
			errChan <- err
		}
	}()

	time.Sleep(1 * time.Second)

	select {
	case err := <-errChan:
		t.Fatalf("server failed to start: %v", err)
	default:
		testMetricsEndpoint(t, port)
		cancel()
	}
}

func getFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port, nil
}

func testMetricsEndpoint(t *testing.T, port int) {
	url := fmt.Sprintf("http://localhost:%d/metrics", port)

	client := &http.Client{Timeout: 3 * time.Second}

	var resp *http.Response
	var err error
	for i := 0; i < 20; i++ {
		resp, err = client.Get(url)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		t.Fatalf("failed to reach /metrics endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status code: got %d, want %d, body: %s",
			resp.StatusCode, http.StatusOK, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if len(body) == 0 {
		t.Fatal("expected non-empty response from /metrics endpoint")
	}

	t.Logf("Successfully received response from /metrics (%d bytes)", len(body))
}
