package analyzer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

func TestAnthropicAnalyzer_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			t.Error("expected x-api-key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("unexpected anthropic-version: %s", r.Header.Get("anthropic-version"))
		}

		resp := map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": `{"score": 8, "reason": "Active project with good docs"}`},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := NewAnthropicAnalyzer("test-key", "claude-haiku-4-5-20251001")
	a.baseURL = srv.URL

	summary := domain.RepoSummary{
		Name:        "test-repo",
		Description: "A test repo",
		Languages:   []string{"Go"},
		FileCount:   10,
		SizeKB:      100,
	}

	report, err := a.Analyze(context.Background(), summary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Score != 8 {
		t.Errorf("expected score 8, got %d", report.Score)
	}
	if report.Reason != "Active project with good docs" {
		t.Errorf("unexpected reason: %s", report.Reason)
	}
	if report.Provider != "anthropic" {
		t.Errorf("expected provider anthropic, got %s", report.Provider)
	}
	if report.Repo != "test-repo" {
		t.Errorf("expected repo test-repo, got %s", report.Repo)
	}
}

func TestAnthropicAnalyzer_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	a := NewAnthropicAnalyzer("test-key", "claude-haiku-4-5-20251001")
	a.baseURL = srv.URL

	_, err := a.Analyze(context.Background(), domain.RepoSummary{Name: "repo"})
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestAnthropicAnalyzer_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": "not valid json at all"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := NewAnthropicAnalyzer("test-key", "claude-haiku-4-5-20251001")
	a.baseURL = srv.URL

	_, err := a.Analyze(context.Background(), domain.RepoSummary{Name: "repo"})
	if err == nil {
		t.Fatal("expected error for invalid JSON in response")
	}
}

func TestAnthropicAnalyzer_MarkdownCodeBlock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": "```json\n{\"score\": 5, \"reason\": \"Wrapped in markdown\"}\n```"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := NewAnthropicAnalyzer("test-key", "claude-haiku-4-5-20251001")
	a.baseURL = srv.URL

	report, err := a.Analyze(context.Background(), domain.RepoSummary{Name: "repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Score != 5 {
		t.Errorf("expected score 5, got %d", report.Score)
	}
}
