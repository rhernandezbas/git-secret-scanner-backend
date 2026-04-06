package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

// newHTTPClient returns an HTTP client tuned for Docker/Alpine environments.
// Disables HTTP/2 and sets explicit timeouts to avoid TLS handshake timeouts
// caused by MTU mismatches in container networking.
func newHTTPClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		DisableCompression:  false,
		ForceAttemptHTTP2:   false, // disable HTTP/2 to avoid MTU fragmentation issues
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   30 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   120 * time.Second,
	}
}

type GitHubProvider struct {
	client *github.Client
}

func NewGitHubProvider(token string) *GitHubProvider {
	httpClient := newHTTPClient()
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient.Transport = &oauth2.Transport{
			Source: ts,
			Base:   httpClient.Transport,
		}
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
