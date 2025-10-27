package integration

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/RyRose/uplog/test/testutil"
	_ "github.com/mattn/go-sqlite3"
)

// TestIntegration_IndexPages tests all index page endpoints.
func TestIntegration_IndexPages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := testutil.Setup(t)
	defer srv.Cancel()

	testCases := []struct {
		name     string
		endpoint string
	}{
		{"main index", "/"},
		{"data index", "/data/"},
		{"data with tabs", "/data/lift/movement"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := srv.Get(t, tc.endpoint)
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

			// Verify it's HTML
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			// All index pages should have html element
			if doc.Find("html").Length() == 0 {
				t.Errorf("expected html element in index page")
			}
		})
	}
}

// TestIntegration_MainViewEndpoints tests main view tab endpoints.
func TestIntegration_MainViewEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := testutil.Setup(t)
	defer srv.Cancel()

	testCases := []struct {
		name     string
		endpoint string
	}{
		{"main tab", "/view/tabs/main"},
		{"lift groups", "/view/liftgroups"},
		{"progress table", "/view/progresstable"},
		{"lift select", "/view/liftselect"},
		{"side weight select", "/view/sideweightselect"},
		{"progress form", "/view/progressform"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := srv.Get(t, tc.endpoint)
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("GET %s: unexpected status code: got %d, want %d, body: %s",
					tc.endpoint, resp.StatusCode, http.StatusOK, string(body))
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			// Verify it returns some content (even if just empty HTML elements)
			// Length check is lenient since some views may return minimal HTML when empty
			if len(body) == 0 {
				t.Errorf("expected some response from %s", tc.endpoint)
			}
		})
	}
}
