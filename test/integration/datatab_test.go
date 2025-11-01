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

// TestIntegration_DataTabEndpoints tests data tab navigation endpoints.
func TestIntegration_DataTabEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := testutil.Setup(t)
	defer srv.Cancel()

	testCases := []struct {
		name     string
		endpoint string
	}{
		{"data tab root", "/view/tabs/data/"},
		{"data tab with indices", "/view/tabs/data/0/1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := srv.Get(t, tc.endpoint)
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

			// Verify HTML structure
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			// Data tab should have some content
			if doc.Text() == "" {
				t.Error("expected non-empty data tab content")
			}
		})
	}
}
