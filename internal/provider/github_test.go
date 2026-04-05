package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v60/github"
)

func newTestGitHubProvider(serverURL string) *GitHubProvider {
	client, _ := github.NewClient(nil).WithEnterpriseURLs(serverURL+"/", serverURL+"/")
	return &GitHubProvider{client: client}
}

func TestGitHubProvider_ListPublicRepos(t *testing.T) {
	repos := []map[string]interface{}{
		{
			"name":             "repo-one",
			"full_name":        "user/repo-one",
			"description":      "First repo",
			"clone_url":        "https://github.com/user/repo-one.git",
			"size":             42,
			"language":         "Go",
			"stargazers_count": 10,
		},
		{
			"name":             "repo-two",
			"full_name":        "user/repo-two",
			"description":      "Second repo",
			"clone_url":        "https://github.com/user/repo-two.git",
			"size":             7,
			"language":         "Python",
			"stargazers_count": 3,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repos)
	}))
	defer srv.Close()

	p := newTestGitHubProvider(srv.URL)
	got, err := p.ListPublicRepos(context.Background(), "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(got))
	}

	if got[0].Name != "repo-one" {
		t.Errorf("expected repo-one, got %s", got[0].Name)
	}
	if got[0].Stars != 10 {
		t.Errorf("expected 10 stars, got %d", got[0].Stars)
	}
	if got[0].Provider != "github" {
		t.Errorf("expected provider github, got %s", got[0].Provider)
	}
	if got[1].Language != "Python" {
		t.Errorf("expected Python, got %s", got[1].Language)
	}
}

func TestGitHubProvider_EmptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
	}))
	defer srv.Close()

	p := newTestGitHubProvider(srv.URL)
	got, err := p.ListPublicRepos(context.Background(), "emptyuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 repos, got %d", len(got))
	}
}

func TestGitHubProvider_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	p := newTestGitHubProvider(srv.URL)
	_, err := p.ListPublicRepos(context.Background(), "nouser")
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
}

// TestGitHubProvider_Pagination documents pagination behavior.
// The provider follows NextPage links until resp.NextPage == 0,
// accumulating all repos across pages. This is validated by the
// loop in ListPublicRepos that checks resp.NextPage.
func TestGitHubProvider_Pagination(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		page++
		if page == 1 {
			// Simulate Link header for next page — go-github parses this
			w.Header().Set("Link", `<http://`+r.Host+r.URL.Path+`?page=2>; rel="next"`)
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{"name": "page1-repo", "full_name": "u/page1-repo"},
			})
		} else {
			json.NewEncoder(w).Encode([]map[string]interface{}{
				{"name": "page2-repo", "full_name": "u/page2-repo"},
			})
		}
	}))
	defer srv.Close()

	p := newTestGitHubProvider(srv.URL)
	got, err := p.ListPublicRepos(context.Background(), "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 repos across 2 pages, got %d", len(got))
	}
}
