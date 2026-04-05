package cloner

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

// mockGitRunner records calls and returns a configured error.
type mockGitRunner struct {
	err      error
	called   bool
	lastPath string
}

func (m *mockGitRunner) PlainCloneContext(_ context.Context, path string, _ bool, _ *gogit.CloneOptions) error {
	m.called = true
	m.lastPath = path
	return m.err
}

func TestClone_EmptyCloneURL_ReturnsError(t *testing.T) {
	t.Parallel()

	c := New(t.TempDir(), 100)
	repo := domain.RepoInfo{Name: "myrepo", CloneURL: ""}

	_, err := c.Clone(context.Background(), repo)
	if err == nil {
		t.Fatal("expected error for empty CloneURL, got nil")
	}
	if !strings.Contains(err.Error(), "clone URL is empty") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestClone_ExceedsMaxSize_ReturnsError(t *testing.T) {
	t.Parallel()

	// maxRepoSizeMB = 1, repo size = 2048 KB (2 MB) → exceeds limit
	c := New(t.TempDir(), 1)
	repo := domain.RepoInfo{
		Name:     "bigrepo",
		CloneURL: "https://github.com/example/bigrepo.git",
		SizeKB:   2048,
	}

	_, err := c.Clone(context.Background(), repo)
	if err == nil {
		t.Fatal("expected error for oversized repo, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds max size") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestClone_Success_ReturnsTempDirPath(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	runner := &mockGitRunner{err: nil}
	c := newWithRunner(tmp, 100, runner)

	repo := domain.RepoInfo{
		Name:     "myrepo",
		CloneURL: "https://github.com/example/myrepo.git",
		SizeKB:   512,
	}

	got, err := c.Clone(context.Background(), repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !runner.called {
		t.Fatal("expected git runner to be called")
	}
	want := filepath.Join(tmp, "gss-myrepo")
	if got != want {
		t.Fatalf("expected path %q, got %q", want, got)
	}
}

func TestClone_GitError_ReturnsWrappedError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	gitErr := errors.New("authentication required")
	runner := &mockGitRunner{err: gitErr}
	c := newWithRunner(tmp, 100, runner)

	repo := domain.RepoInfo{
		Name:     "privaterepo",
		CloneURL: "https://github.com/example/privaterepo.git",
		SizeKB:   100,
	}

	_, err := c.Clone(context.Background(), repo)
	if err == nil {
		t.Fatal("expected error from git runner, got nil")
	}
	if !errors.Is(err, gitErr) {
		t.Fatalf("expected wrapped git error, got: %v", err)
	}
}

func TestDelete_EmptyPath_ReturnsNil(t *testing.T) {
	t.Parallel()

	if err := Delete(""); err != nil {
		t.Fatalf("expected nil for empty path, got: %v", err)
	}
}

func TestDelete_ExistingDir_RemovesIt(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	if err := Delete(tmp); err != nil {
		t.Fatalf("unexpected error deleting temp dir: %v", err)
	}
}
