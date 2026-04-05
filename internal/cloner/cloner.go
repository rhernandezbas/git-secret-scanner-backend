package cloner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

// GitRunner abstracts the actual git clone operation for testability.
type GitRunner interface {
	PlainCloneContext(ctx context.Context, path string, isBare bool, o *gogit.CloneOptions) error
}

// defaultGitRunner delegates to go-git directly.
type defaultGitRunner struct{}

func (defaultGitRunner) PlainCloneContext(ctx context.Context, path string, isBare bool, o *gogit.CloneOptions) error {
	_, err := gogit.PlainCloneContext(ctx, path, isBare, o)
	return err
}

type GitCloner struct {
	tempDir       string
	maxRepoSizeMB int
	runner        GitRunner
}

func New(tempDir string, maxRepoSizeMB int) *GitCloner {
	return &GitCloner{
		tempDir:       tempDir,
		maxRepoSizeMB: maxRepoSizeMB,
		runner:        defaultGitRunner{},
	}
}

// newWithRunner is used by tests to inject a mock runner.
func newWithRunner(tempDir string, maxRepoSizeMB int, runner GitRunner) *GitCloner {
	return &GitCloner{
		tempDir:       tempDir,
		maxRepoSizeMB: maxRepoSizeMB,
		runner:        runner,
	}
}

func (c *GitCloner) Clone(ctx context.Context, repo domain.RepoInfo) (string, error) {
	if repo.CloneURL == "" {
		return "", fmt.Errorf("clone URL is empty for repo %s", repo.Name)
	}

	if repo.SizeKB > c.maxRepoSizeMB*1024 {
		return "", fmt.Errorf("repo %s exceeds max size (%d MB)", repo.Name, c.maxRepoSizeMB)
	}

	destPath := filepath.Join(c.tempDir, fmt.Sprintf("gss-%s", repo.Name))

	// Remove if exists from a previous failed run
	_ = os.RemoveAll(destPath)

	err := c.runner.PlainCloneContext(ctx, destPath, false, &gogit.CloneOptions{
		URL:           repo.CloneURL,
		Depth:         1,
		SingleBranch:  true,
		ReferenceName: plumbing.HEAD,
		Progress:      nil,
	})
	if err != nil {
		_ = os.RemoveAll(destPath)
		return "", fmt.Errorf("cloning %s: %w", repo.Name, err)
	}

	return destPath, nil
}
