package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

// ProviderFactory creates a RepoProvider for the given provider name.
type ProviderFactory func(providerName string) (domain.RepoProvider, error)

type ScanService struct {
	providerFactory ProviderFactory
	cloner          domain.Cloner
	scanner         domain.Scanner
	repo            domain.FindingsRepository
	broadcaster     domain.EventBroadcaster
	analyzer        domain.Analyzer // nil when AI_ANALYSIS_ENABLED=false
	aiEnabled       bool
}

func NewScanService(
	providerFactory ProviderFactory,
	cloner domain.Cloner,
	scanner domain.Scanner,
	repo domain.FindingsRepository,
	broadcaster domain.EventBroadcaster,
	analyzer domain.Analyzer,
	aiEnabled bool,
) *ScanService {
	return &ScanService{
		providerFactory: providerFactory,
		cloner:          cloner,
		scanner:         scanner,
		repo:            repo,
		broadcaster:     broadcaster,
		analyzer:        analyzer,
		aiEnabled:       aiEnabled,
	}
}

func (s *ScanService) Run(ctx context.Context, req domain.ScanRequest) error {
	// 1. Broadcast scan_start
	s.broadcast(domain.EventScanStart, "", fmt.Sprintf("Starting scan for user: %s on %s", req.Username, req.Provider), nil)

	// 2. Resolve provider for this request
	repoProvider, err := s.providerFactory(req.Provider)
	if err != nil {
		s.broadcast(domain.EventError, "", fmt.Sprintf("Failed to build provider: %v", err), nil)
		return err
	}

	// 3. List repos
	repos, err := repoProvider.ListPublicRepos(ctx, req.Username)
	if err != nil {
		s.broadcast(domain.EventError, "", fmt.Sprintf("Failed to list repos: %v", err), nil)
		return err
	}

	s.broadcast(domain.EventScanStart, "", fmt.Sprintf("Found %d public repos", len(repos)), map[string]int{"count": len(repos)})

	// 3. Process each repo
	for _, repo := range repos {
		s.processRepo(ctx, repo)
	}

	// 4. Broadcast scan_complete
	s.broadcast(domain.EventScanComplete, "", fmt.Sprintf("Scan complete for %s — %d repos processed", req.Username, len(repos)), nil)
	return nil
}

func (s *ScanService) processRepo(ctx context.Context, repo domain.RepoInfo) {
	start := time.Now()
	s.broadcast(domain.EventRepoStart, repo.Name, fmt.Sprintf("Cloning %s...", repo.FullName), nil)

	// Clone
	localPath, err := s.cloner.Clone(ctx, repo)
	if err != nil {
		s.broadcast(domain.EventError, repo.Name, fmt.Sprintf("Clone failed: %v", err), nil)
		return
	}

	// Always delete when done
	defer func() {
		_ = os.RemoveAll(localPath)
	}()

	// Scan
	s.broadcast(domain.EventFileScanned, repo.Name, "Scanning for secrets...", nil)
	findings, err := s.scanner.Scan(ctx, repo, localPath)
	if err != nil {
		s.broadcast(domain.EventError, repo.Name, fmt.Sprintf("Scan error: %v", err), nil)
		// still continue to save partial results
	}

	// Broadcast each finding
	for _, f := range findings {
		s.broadcast(domain.EventFindingFound, repo.Name, fmt.Sprintf("[%s] %s in %s:%d", f.Severity, f.PatternName, f.FilePath, f.Line), f)
	}

	result := domain.ScanResult{
		Repo:      repo,
		Findings:  findings,
		Duration:  time.Since(start).String(),
		ScannedAt: time.Now(),
	}

	// AI Analysis (feature flag)
	if s.aiEnabled && s.analyzer != nil {
		s.broadcast(domain.EventAIAnalysisStart, repo.Name, "Running AI analysis...", nil)
		summary := buildRepoSummary(repo, localPath)
		report, err := s.analyzer.Analyze(ctx, summary)
		if err != nil {
			s.broadcast(domain.EventError, repo.Name, fmt.Sprintf("AI analysis failed: %v", err), nil)
		} else {
			result.UtilityReport = &report
			s.broadcast(domain.EventAIAnalysisDone, repo.Name, fmt.Sprintf("AI score: %d/10 — %s", report.Score, report.Reason), report)
		}
	}

	// Persist
	if err := s.repo.Save(ctx, result); err != nil {
		s.broadcast(domain.EventError, repo.Name, fmt.Sprintf("Failed to save findings: %v", err), nil)
	}

	s.broadcast(domain.EventRepoDone, repo.Name,
		fmt.Sprintf("Done — %d findings in %s", len(findings), time.Since(start).Round(time.Millisecond)),
		map[string]interface{}{"findings_count": len(findings), "duration": result.Duration},
	)
}

func (s *ScanService) broadcast(eventType domain.EventType, repo, message string, payload interface{}) {
	s.broadcaster.Broadcast(domain.ProgressEvent{
		Type:    eventType,
		Repo:    repo,
		Message: message,
		Payload: payload,
	})
}

func buildRepoSummary(repo domain.RepoInfo, localPath string) domain.RepoSummary {
	summary := domain.RepoSummary{
		Name:        repo.Name,
		Description: repo.Description,
		SizeKB:      repo.SizeKB,
	}

	// Collect top-level file names
	entries, err := os.ReadDir(localPath)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				summary.TopFiles = append(summary.TopFiles, e.Name())
			}
			if len(summary.TopFiles) >= 10 {
				break
			}
		}
	}

	// Count files
	_ = filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			summary.FileCount++
		}
		return nil
	})

	// README excerpt
	for _, name := range []string{"README.md", "readme.md", "README.txt", "README"} {
		data, err := os.ReadFile(filepath.Join(localPath, name))
		if err == nil {
			excerpt := string(data)
			if len(excerpt) > 500 {
				excerpt = excerpt[:500]
			}
			summary.ReadmeExcerpt = excerpt
			break
		}
	}

	return summary
}
