package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/handler"
)

func TestFindingsGetAll_ReturnsJSONArray(t *testing.T) {
	repo := &mockFindingsRepo{
		results: []domain.ScanResult{
			{
				Repo:      domain.RepoInfo{Name: "repo1", FullName: "user/repo1", Provider: "github"},
				Findings:  []domain.Finding{},
				Duration:  "1s",
				ScannedAt: time.Now(),
			},
		},
	}
	h := handler.NewFindingsHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/findings", nil)
	rr := httptest.NewRecorder()

	h.GetAll(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var results []domain.ScanResult
	if err := json.NewDecoder(rr.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Repo.Name != "repo1" {
		t.Errorf("expected repo name 'repo1', got %q", results[0].Repo.Name)
	}
}

func TestFindingsGetAll_ReturnsEmptyArray_NotNull(t *testing.T) {
	repo := &mockFindingsRepo{results: nil}
	h := handler.NewFindingsHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/findings", nil)
	rr := httptest.NewRecorder()

	h.GetAll(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	// Must not be "null\n" — must be an array
	var results []domain.ScanResult
	if err := json.NewDecoder(rr.Body).Decode(&results); err != nil {
		// Re-try with the already-read body string
		if err2 := json.Unmarshal([]byte(body), &results); err2 != nil {
			t.Fatalf("failed to decode response body %q: %v", body, err2)
		}
	}
	if results == nil {
		t.Error("expected empty array, got null")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
