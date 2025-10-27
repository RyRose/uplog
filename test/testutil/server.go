package testutil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/RyRose/uplog/internal/service"
	_ "github.com/mattn/go-sqlite3"
)

// Server represents a test server instance.
type Server struct {
	Port   int
	cancel context.CancelFunc
	errCh  chan error
}

// Setup creates and starts a test server instance.
// The server will be automatically stopped when the test completes.
func Setup(t *testing.T) *Server {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	port, err := getFreePort()
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}

	os.Setenv("PORT", fmt.Sprintf("%d", port))
	os.Setenv("DATABASE_PATH", dbPath)
	os.Setenv("DEBUG", "false")
	t.Cleanup(func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_PATH")
		os.Unsetenv("DEBUG")
	})

	// Change to repo root so Lua's require() can find config modules
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("failed to get repo root: %v", err)
	}

	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change to repo root: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	errChan := make(chan error, 1)
	go func() {
		if err := service.Run(ctx, "./config/main.lua"); err != nil {
			errChan <- err
		}
	}()

	// Wait for server to be ready by polling health endpoint
	if err := waitForServer(port, 5*time.Second); err != nil {
		select {
		case serverErr := <-errChan:
			t.Fatalf("server failed to start: %v", serverErr)
		default:
			t.Fatalf("server did not become ready: %v", err)
		}
	}

	return &Server{
		Port:   port,
		cancel: cancel,
		errCh:  errChan,
	}
}

// Get makes an HTTP GET request to the given path and returns the response.
// The path should start with a slash (e.g., "/health").
// The method will retry on connection errors with exponential backoff.
func (s *Server) Get(t *testing.T, path string) *http.Response {
	t.Helper()
	url := fmt.Sprintf("http://localhost:%d%s", s.Port, path)
	client := &http.Client{Timeout: 3 * time.Second}

	var resp *http.Response
	var err error
	for range 20 {
		resp, err = client.Get(url)
		if err == nil {
			return resp
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("failed to reach endpoint %s: %v", path, err)
	return nil
}

// Cancel stops the server.
func (s *Server) Cancel() {
	s.cancel()
}

func waitForServer(port int, timeout time.Duration) error {
	url := fmt.Sprintf("http://localhost:%d/health", port)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for server to be ready")
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
