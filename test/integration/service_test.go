package integration

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/RyRose/uplog/test/testutil"
	_ "github.com/mattn/go-sqlite3"
)

func TestIntegration_Endpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := testutil.Setup(t)
	defer srv.Cancel()

	t.Run("Health", func(t *testing.T) {
		resp := srv.Get(t, "/health")
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("unexpected status code: got %d, want %d, body: %s",
				resp.StatusCode, http.StatusOK, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != "OK" {
			t.Errorf("expected 'OK', got %q", string(body))
		}
	})

	t.Run("Metrics", func(t *testing.T) {
		resp := srv.Get(t, "/metrics")
		defer func() { _ = resp.Body.Close() }()

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
	})

	t.Run("IndexPage", func(t *testing.T) {
		resp := srv.Get(t, "/")
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("GET /: unexpected status code: got %d, want %d, body: %s",
				resp.StatusCode, http.StatusOK, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if len(body) == 0 {
			t.Fatal("expected non-empty response from index page")
		}

		bodyStr := string(body)
		if !strings.Contains(bodyStr, "<!DOCTYPE html>") && !strings.Contains(bodyStr, "<html") {
			t.Errorf("expected HTML content, got: %s", bodyStr[:min(200, len(bodyStr))])
		}
	})

	t.Run("DataPage", func(t *testing.T) {
		resp := srv.Get(t, "/data/")
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("GET /data/: unexpected status code: got %d, want %d, body: %s",
				resp.StatusCode, http.StatusOK, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if len(body) == 0 {
			t.Fatal("expected non-empty response from data page")
		}
	})

	t.Run("SwaggerDocs", func(t *testing.T) {
		resp := srv.Get(t, "/docs/swagger.json")
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("GET /docs/swagger.json: unexpected status code: got %d, want %d, body: %s",
				resp.StatusCode, http.StatusOK, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if len(body) == 0 {
			t.Fatal("expected non-empty swagger.json")
		}

		bodyStr := string(body)
		if !strings.Contains(bodyStr, "swagger") && !strings.Contains(bodyStr, "openapi") {
			t.Errorf("expected swagger/openapi content in swagger.json")
		}
	})

	t.Run("ViewEndpoints", func(t *testing.T) {
		tests := []struct {
			name     string
			path     string
			wantCode int
		}{
			{"main tab", "/view/tabs/main", http.StatusOK},
			{"data tab", "/view/tabs/data/", http.StatusOK},
			{"lift groups", "/view/liftgroups", http.StatusOK},
			{"progress table", "/view/progresstable", http.StatusOK},
			{"routine table", "/view/routinetable", http.StatusOK},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp := srv.Get(t, tt.path)
				defer func() { _ = resp.Body.Close() }()

				if resp.StatusCode != tt.wantCode {
					body, _ := io.ReadAll(resp.Body)
					t.Errorf("GET %s: got status %d, want %d, body: %s",
						tt.path, resp.StatusCode, tt.wantCode, string(body))
				}
			})
		}
	})

	t.Run("DataViewEndpoints", func(t *testing.T) {
		dataEndpoints := []string{
			"/view/data/lift",
			"/view/data/movement",
			"/view/data/muscle",
			"/view/data/routine",
			"/view/data/side_weight",
			"/view/data/template_variable",
			"/view/data/workout",
			"/view/data/progress",
			"/view/data/lift_muscle_mapping",
			"/view/data/lift_workout_mapping",
			"/view/data/routine_workout_mapping",
			"/view/data/subworkout",
			"/view/data/lift_group",
		}

		for _, endpoint := range dataEndpoints {
			t.Run(endpoint, func(t *testing.T) {
				resp := srv.Get(t, endpoint)
				defer func() { _ = resp.Body.Close() }()

				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					t.Errorf("GET %s: got status %d, want %d, body: %s",
						endpoint, resp.StatusCode, http.StatusOK, string(body))
				}
			})
		}
	})
}
