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

// TestIntegration_ScheduleEndpoints tests schedule view endpoints.
func TestIntegration_ScheduleEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := testutil.Setup(t)
	defer srv.Cancel()

	t.Run("schedule tab", func(t *testing.T) {
		resp := srv.Get(t, "/view/tabs/schedule")
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

		// Parse HTML and verify structure
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			t.Fatalf("failed to parse HTML: %v", err)
		}

		// Verify table exists
		tables := doc.Find("table")
		if tables.Length() == 0 {
			t.Error("expected table element in schedule view")
		}
	})
}
