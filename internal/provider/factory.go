package provider

import (
	"fmt"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

func New(providerName, githubToken, gitlabToken string) (domain.RepoProvider, error) {
	switch providerName {
	case "github":
		return NewGitHubProvider(githubToken), nil
	case "gitlab":
		p, err := NewGitLabProvider(gitlabToken)
		if err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s (use github or gitlab)", providerName)
	}
}
