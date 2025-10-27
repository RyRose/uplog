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

// TestIntegration_RawDataViewStructure tests that all raw data view GET endpoints return valid HTML with expected structure.
func TestIntegration_RawDataViewStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := testutil.Setup(t)
	defer srv.Cancel()

	testCases := []struct {
		name     string
		endpoint string
		wantCols int // minimum expected columns
	}{
		// Single entity tables
		{"lift table", "/view/data/lift", 3},
		{"movement table", "/view/data/movement", 2},
		{"muscle table", "/view/data/muscle", 2},
		{"routine table", "/view/data/routine", 3},
		{"workout table", "/view/data/workout", 1},
		{"schedule list table", "/view/data/schedule_list", 1},
		{"side weight table", "/view/data/side_weight", 3},
		{"template variable table", "/view/data/template_variable", 2},
		{"schedule table", "/view/data/schedule", 2},
		{"progress table", "/view/data/progress", 5},
		{"lift group table", "/view/data/lift_group", 1},

		// Mapping/relationship tables
		{"lift muscle mapping", "/view/data/lift_muscle_mapping", 3},
		{"lift workout mapping", "/view/data/lift_workout_mapping", 2},
		{"routine workout mapping", "/view/data/routine_workout_mapping", 2},
		{"subworkout mapping", "/view/data/subworkout", 2},
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

			// Parse HTML to ensure it's valid
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			// Verify table exists
			tables := doc.Find("table")
			if tables.Length() < 1 {
				t.Errorf("expected at least one table in %s", tc.name)
			}

			// Verify table has header row with expected columns
			headerCells := doc.Find("thead th")
			if headerCells.Length() < tc.wantCols {
				t.Errorf("%s: expected at least %d header columns, got %d",
					tc.name, tc.wantCols, headerCells.Length())
			}

			// Verify table has tbody (even if empty)
			tbody := doc.Find("tbody")
			if tbody.Length() < 1 {
				t.Errorf("%s: expected tbody element", tc.name)
			}
		})
	}
}

// TestIntegration_DefaultDataExists tests that the default migration data is present.
func TestIntegration_DefaultDataExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := testutil.Setup(t)
	defer srv.Cancel()

	// The default migration should insert some lifts
	resp := srv.Get(t, "/view/data/lift")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	// Verify there are some rows in the table (default data)
	tableRows := doc.Find("tbody tr")
	if tableRows.Length() == 0 {
		t.Error("expected default migration to insert at least some lift data")
	}
}

// TestIntegration_EmptyTables tests that empty tables render correctly.
func TestIntegration_EmptyTables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := testutil.Setup(t)
	defer srv.Cancel()

	testCases := []struct {
		name     string
		endpoint string
	}{
		{"lift table", "/view/data/lift"},
		{"movement table", "/view/data/movement"},
		{"muscle table", "/view/data/muscle"},
		{"routine table", "/view/data/routine"},
		{"workout table", "/view/data/workout"},
		{"progress table", "/view/progresstable"},
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

			// Parse HTML to ensure it's valid
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			// Verify table exists (even if empty)
			tables := doc.Find("table")
			if tables.Length() < 1 {
				t.Errorf("expected at least one table in %s", tc.name)
			}
		})
	}
}
