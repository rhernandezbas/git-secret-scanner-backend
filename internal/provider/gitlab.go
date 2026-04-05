package provider

import (
	"context"
	"fmt"

	"github.com/xanzy/go-gitlab"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

type GitLabProvider struct {
	client *gitlab.Client
}

func NewGitLabProvider(token string) (*GitLabProvider, error) {
	client, err := gitlab.NewClient(token)
	if err != nil {
		return nil, err
	}
	return &GitLabProvider{client: client}, nil
}

func (p *GitLabProvider) ListPublicRepos(ctx context.Context, username string) ([]domain.RepoInfo, error) {
	var allRepos []domain.RepoInfo

	visibility := gitlab.PublicVisibility
	opts := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
		Visibility:  &visibility,
	}

	for {
		projects, resp, err := p.client.Projects.ListUserProjects(username, opts)
		if err != nil {
			return nil, fmt.Errorf("listing gitlab repos for %s: %w", username, err)
		}

		for _, proj := range projects {
			var sizeKB int
			if proj.Statistics != nil {
				sizeKB = int(proj.Statistics.RepositorySize / 1024)
			}
			allRepos = append(allRepos, domain.RepoInfo{
				Name:        proj.Name,
				FullName:    proj.PathWithNamespace,
				Description: proj.Description,
				CloneURL:    proj.HTTPURLToRepo,
				SizeKB:      sizeKB,
				Language:    "",
				Stars:       proj.StarCount,
				Provider:    "gitlab",
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}
