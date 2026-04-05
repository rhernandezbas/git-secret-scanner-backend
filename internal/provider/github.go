package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

type GitHubProvider struct {
	client *github.Client
}

func NewGitHubProvider(token string) *GitHubProvider {
	var httpClient *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient = oauth2.NewClient(context.Background(), ts)
	}
	return &GitHubProvider{client: github.NewClient(httpClient)}
}

func (p *GitHubProvider) ListPublicRepos(ctx context.Context, username string) ([]domain.RepoInfo, error) {
	var allRepos []domain.RepoInfo
	opts := &github.RepositoryListByUserOptions{
		Type:        "public",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := p.client.Repositories.ListByUser(ctx, username, opts)
		if err != nil {
			return nil, fmt.Errorf("listing repos for %s: %w", username, err)
		}

		for _, r := range repos {
			allRepos = append(allRepos, domain.RepoInfo{
				Name:        r.GetName(),
				FullName:    r.GetFullName(),
				Description: r.GetDescription(),
				CloneURL:    r.GetCloneURL(),
				SizeKB:      r.GetSize(),
				Language:    r.GetLanguage(),
				Stars:       r.GetStargazersCount(),
				Provider:    "github",
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}
