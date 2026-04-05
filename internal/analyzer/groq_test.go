package analyzer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

func TestGroqAnalyzer_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer groq-key" {
			t.Errorf("unexpected auth header: %s", auth)
		}
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{
					"content": `{"score": 6, "reason": "Decent open-source tool"}`,
				}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := NewGroqAnalyzer("groq-key", "llama-3.1-8b-instant")
	a.baseURL = srv.URL

	report, err := a.Analyze(context.Background(), domain.RepoSummary{Name: "groq-repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Score != 6 {
		t.Errorf("expected score 6, got %d", report.Score)
	}
	if report.Provider != "groq" {
		t.Errorf("expected provider groq, got %s", report.Provider)
	}
	if report.Model != "llama-3.1-8b-instant" {
		t.Errorf("unexpected model: %s", report.Model)
	}
}

func TestGroqAnalyzer_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	a := NewGroqAnalyzer("bad-key", "llama-3.1-8b-instant")
	a.baseURL = srv.URL

	_, err := a.Analyze(context.Background(), domain.RepoSummary{Name: "repo"})
	if err == nil {
		t.Fatal("expected error for HTTP 401")
	}
}

func TestGroqAnalyzer_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{
					"content": "not json",
				}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := NewGroqAnalyzer("groq-key", "llama-3.1-8b-instant")
	a.baseURL = srv.URL

	_, err := a.Analyze(context.Background(), domain.RepoSummary{Name: "repo"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGroqAnalyzer_MarkdownCodeBlock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{
					"content": "```json\n{\"score\": 9, \"reason\": \"Excellent utility\"}\n```",
				}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := NewGroqAnalyzer("groq-key", "llama-3.1-8b-instant")
	a.baseURL = srv.URL

	report, err := a.Analyze(context.Background(), domain.RepoSummary{Name: "repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Score != 9 {
		t.Errorf("expected score 9, got %d", report.Score)
	}
}
