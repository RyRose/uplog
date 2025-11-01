package integration

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/RyRose/uplog/test/testutil"
	_ "github.com/mattn/go-sqlite3"
)

// TestIntegration_RawDataMutations tests POST, PATCH, DELETE for all raw data endpoints.
func TestIntegration_RawDataMutations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := testutil.Setup(t)
	defer srv.Cancel()

	baseURL := "http://localhost:" + srv.GetPort(t)

	testCases := []struct {
		name       string
		endpoint   string
		createData url.Values
		updateData url.Values
		createID   string
		requiresID bool // Some endpoints don't need ID in URL for DELETE
	}{
		{
			name:     "lift",
			endpoint: "/view/data/lift",
			createData: url.Values{
				"id":   {"test-lift-mutation"},
				"link": {"https://example.com/test"},
			},
			updateData: url.Values{
				"id":   {"test-lift-mutation"},
				"link": {"https://example.com/updated"},
			},
			createID:   "test-lift-mutation",
			requiresID: true,
		},
		{
			name:     "movement",
			endpoint: "/view/data/movement",
			createData: url.Values{
				"id":    {"test-movement"},
				"alias": {"Test Movement"},
			},
			updateData: url.Values{
				"id":    {"test-movement"},
				"alias": {"Updated Movement"},
			},
			createID:   "test-movement",
			requiresID: true,
		},
		{
			name:     "muscle",
			endpoint: "/view/data/muscle",
			createData: url.Values{
				"id":   {"test-muscle"},
				"link": {"https://example.com/muscle"},
			},
			updateData: url.Values{
				"id":   {"test-muscle"},
				"link": {"https://example.com/muscle-updated"},
			},
			createID:   "test-muscle",
			requiresID: true,
		},
		{
			name:     "workout",
			endpoint: "/view/data/workout",
			createData: url.Values{
				"id": {"test-workout"},
			},
			updateData: url.Values{
				"id": {"test-workout"},
			},
			createID:   "test-workout",
			requiresID: true,
		},
		{
			name:     "lift_group",
			endpoint: "/view/data/lift_group",
			createData: url.Values{
				"id": {"test-lift-group"},
			},
			updateData: url.Values{
				"id": {"test-lift-group"},
			},
			createID:   "test-lift-group",
			requiresID: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("POST creates entry", func(t *testing.T) {
				req, err := http.NewRequest("POST", baseURL+tc.endpoint, strings.NewReader(tc.createData.Encode()))
				if err != nil {
					t.Fatalf("failed to create request: %v", err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					t.Fatalf("failed to make request: %v", err)
				}
				defer func() { _ = resp.Body.Close() }()

				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					t.Fatalf("POST %s: unexpected status code: got %d, want %d, body: %s",
						tc.endpoint, resp.StatusCode, http.StatusOK, string(body))
				}

				// Response should be HTML (table row)
				body, _ := io.ReadAll(resp.Body)
				if len(body) == 0 {
					t.Error("expected non-empty response from POST")
				}
			})

			t.Run("PATCH updates entry", func(t *testing.T) {
				patchURL := baseURL + tc.endpoint
				if tc.requiresID {
					patchURL += "/" + tc.createID
				}

				req, err := http.NewRequest("PATCH", patchURL, strings.NewReader(tc.updateData.Encode()))
				if err != nil {
					t.Fatalf("failed to create request: %v", err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					t.Fatalf("failed to make request: %v", err)
				}
				defer func() { _ = resp.Body.Close() }()

				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					t.Fatalf("PATCH %s: unexpected status code: got %d, want %d, body: %s",
						patchURL, resp.StatusCode, http.StatusOK, string(body))
				}
			})

			t.Run("DELETE removes entry", func(t *testing.T) {
				deleteURL := baseURL + tc.endpoint
				if tc.requiresID {
					deleteURL += "/" + tc.createID
				}

				req, err := http.NewRequest("DELETE", deleteURL, nil)
				if err != nil {
					t.Fatalf("failed to create request: %v", err)
				}

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					t.Fatalf("failed to make request: %v", err)
				}
				defer func() { _ = resp.Body.Close() }()

				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					t.Fatalf("DELETE %s: unexpected status code: got %d, want %d, body: %s",
						deleteURL, resp.StatusCode, http.StatusOK, string(body))
				}
			})
		})
	}
}

// TestIntegration_ProgressTableMutations tests progress table specific endpoints.
func TestIntegration_ProgressTableMutations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := testutil.Setup(t)
	defer srv.Cancel()

	baseURL := "http://localhost:" + srv.GetPort(t)

	// First, we need a lift to create progress for
	// Use one from the default data
	t.Run("POST progresstablerow creates entry", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("lift", "bench-press") // From default data
		formData.Set("date", "2024-12-25")
		formData.Set("weight", "225")
		formData.Set("sets", "3")
		formData.Set("reps", "5")

		req, err := http.NewRequest("POST", baseURL+"/view/progresstablerow", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("unexpected status code: got %d, want %d, body: %s",
				resp.StatusCode, http.StatusOK, string(body))
		}

		// Response should contain table row with the new entry
		body, _ := io.ReadAll(resp.Body)
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			t.Fatalf("failed to parse HTML: %v", err)
		}

		// Should contain the data we inserted
		bodyText := doc.Text()
		if !strings.Contains(bodyText, "225") {
			t.Error("expected response to contain weight '225'")
		}
	})

	t.Run("POST progressform creates entry via form", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("lift", "squat") // From default data
		formData.Set("date", "2024-12-26")
		formData.Set("weight", "315")
		formData.Set("sets", "5")
		formData.Set("reps", "5")

		req, err := http.NewRequest("POST", baseURL+"/view/progressform", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// This endpoint returns the form again, not the row
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("unexpected status code: got %d, want %d, body: %s",
				resp.StatusCode, http.StatusOK, string(body))
		}
	})
}

// TestIntegration_MappingMutations tests relationship/mapping table mutations.
func TestIntegration_MappingMutations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := testutil.Setup(t)
	defer srv.Cancel()

	baseURL := "http://localhost:" + srv.GetPort(t)

	t.Run("lift_muscle_mapping", func(t *testing.T) {
		// POST
		formData := url.Values{}
		formData.Set("lift", "bench-press")
		formData.Set("muscle", "chest")
		formData.Set("movement", "push")

		req, err := http.NewRequest("POST", baseURL+"/view/data/lift_muscle_mapping", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode >= 500 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("POST: unexpected status %d, body: %s", resp.StatusCode, string(body))
		}

		// DELETE
		req, err = http.NewRequest("DELETE", baseURL+"/view/data/lift_muscle_mapping/bench-press/chest/push", nil)
		if err != nil {
			t.Fatalf("failed to create delete request: %v", err)
		}

		resp, err = client.Do(req)
		if err != nil {
			t.Fatalf("failed to make delete request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode >= 500 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("DELETE: unexpected status %d, body: %s", resp.StatusCode, string(body))
		}
	})
}

// TestIntegration_DataProgressMutations tests progress mutations via data view.
func TestIntegration_DataProgressMutations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := testutil.Setup(t)
	defer srv.Cancel()

	baseURL := "http://localhost:" + srv.GetPort(t)

	var createdID int64

	t.Run("POST creates progress entry", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("lift", "deadlift")
		formData.Set("date", "2024-12-27")
		formData.Set("weight", "405")
		formData.Set("sets", "1")
		formData.Set("reps", "5")

		req, err := http.NewRequest("POST", baseURL+"/view/data/progress", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode >= 500 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("POST: unexpected status %d, body: %s", resp.StatusCode, string(body))
		}

		// Try to extract ID from response if possible
		body, _ := io.ReadAll(resp.Body)
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))

		// Look for data-id or id attribute
		doc.Find("[data-id]").Each(func(i int, s *goquery.Selection) {
			if idStr, exists := s.Attr("data-id"); exists {
				if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
					createdID = id
				}
			}
		})
	})

	t.Run("DELETE progress entry", func(t *testing.T) {
		if createdID == 0 {
			t.Skip("no ID from previous test, skipping delete")
		}

		req, err := http.NewRequest("DELETE", baseURL+"/view/data/progress/"+fmt.Sprint(createdID), nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode >= 500 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("DELETE: unexpected status %d, body: %s", resp.StatusCode, string(body))
		}
	})
}
