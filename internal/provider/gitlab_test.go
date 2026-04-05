package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xanzy/go-gitlab"
)

func newTestGitLabProvider(serverURL string) *GitLabProvider {
	client, _ := gitlab.NewClient("test-token", gitlab.WithBaseURL(serverURL))
	return &GitLabProvider{client: client}
}

func TestGitLabProvider_ListPublicRepos(t *testing.T) {
	projects := []map[string]interface{}{
		{
			"name":                "gl-repo-one",
			"path_with_namespace": "user/gl-repo-one",
			"description":         "GitLab first repo",
			"http_url_to_repo":    "https://gitlab.com/user/gl-repo-one.git",
			"star_count":          5,
		},
		{
			"name":                "gl-repo-two",
			"path_with_namespace": "user/gl-repo-two",
			"description":         "GitLab second repo",
			"http_url_to_repo":    "https://gitlab.com/user/gl-repo-two.git",
			"star_count":          8,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	}))
	defer srv.Close()

	p := newTestGitLabProvider(srv.URL)
	got, err := p.ListPublicRepos(context.Background(), "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(got))
	}

	if got[0].Name != "gl-repo-one" {
		t.Errorf("expected gl-repo-one, got %s", got[0].Name)
	}
	if got[0].Stars != 5 {
		t.Errorf("expected 5 stars, got %d", got[0].Stars)
	}
	if got[0].Provider != "gitlab" {
		t.Errorf("expected provider gitlab, got %s", got[0].Provider)
	}
	if got[0].FullName != "user/gl-repo-one" {
		t.Errorf("expected user/gl-repo-one, got %s", got[0].FullName)
	}
}

func TestGitLabProvider_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := newTestGitLabProvider(srv.URL)
	_, err := p.ListPublicRepos(context.Background(), "someuser")
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
}
