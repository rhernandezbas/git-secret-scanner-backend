package domain

import "context"

// RepoProvider lists public repos for a given username
type RepoProvider interface {
	ListPublicRepos(ctx context.Context, username string) ([]RepoInfo, error)
}

// Cloner clones a repo to a temp directory and returns the path
type Cloner interface {
	Clone(ctx context.Context, repo RepoInfo) (localPath string, err error)
}

// Scanner scans a directory and returns all findings
type Scanner interface {
	Scan(ctx context.Context, repo RepoInfo, localPath string) ([]Finding, error)
}

// Analyzer analyzes a repo and returns a utility report
type Analyzer interface {
	Analyze(ctx context.Context, summary RepoSummary) (UtilityReport, error)
}

// FindingsRepository persists and retrieves findings
type FindingsRepository interface {
	Save(ctx context.Context, result ScanResult) error
	LoadAll(ctx context.Context) ([]ScanResult, error)
}

// EventBroadcaster broadcasts progress events to connected clients
type EventBroadcaster interface {
	Broadcast(event ProgressEvent)
	Register(client chan ProgressEvent)
	Unregister(client chan ProgressEvent)
}
