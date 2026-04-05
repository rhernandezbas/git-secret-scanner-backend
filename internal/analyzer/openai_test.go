package analyzer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

func TestOpenAIAnalyzer_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", auth)
		}
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{
					"content": `{"score": 7, "reason": "Well-structured project"}`,
				}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := NewOpenAIAnalyzer("test-key", "gpt-4o-mini")
	a.baseURL = srv.URL

	report, err := a.Analyze(context.Background(), domain.RepoSummary{Name: "my-repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Score != 7 {
		t.Errorf("expected score 7, got %d", report.Score)
	}
	if report.Provider != "openai" {
		t.Errorf("expected provider openai, got %s", report.Provider)
	}
}

func TestOpenAIAnalyzer_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer srv.Close()

	a := NewOpenAIAnalyzer("test-key", "gpt-4o-mini")
	a.baseURL = srv.URL

	_, err := a.Analyze(context.Background(), domain.RepoSummary{Name: "repo"})
	if err == nil {
		t.Fatal("expected error for HTTP 429")
	}
}

func TestOpenAIAnalyzer_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{
					"content": "this is not json",
				}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := NewOpenAIAnalyzer("test-key", "gpt-4o-mini")
	a.baseURL = srv.URL

	_, err := a.Analyze(context.Background(), domain.RepoSummary{Name: "repo"})
	if err == nil {
		t.Fatal("expected error for invalid JSON in response")
	}
}

func TestOpenAIAnalyzer_MarkdownCodeBlock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{
					"content": "```\n{\"score\": 3, \"reason\": \"Minimal repo\"}\n```",
				}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := NewOpenAIAnalyzer("test-key", "gpt-4o-mini")
	a.baseURL = srv.URL

	report, err := a.Analyze(context.Background(), domain.RepoSummary{Name: "repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Score != 3 {
		t.Errorf("expected score 3, got %d", report.Score)
	}
}
