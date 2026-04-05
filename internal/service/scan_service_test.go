package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

// --- Mocks ---

type mockProvider struct {
	repos []domain.RepoInfo
	err   error
}

func (m *mockProvider) ListPublicRepos(_ context.Context, _ string) ([]domain.RepoInfo, error) {
	return m.repos, m.err
}

type mockCloner struct {
	path string
	err  error
}

func (m *mockCloner) Clone(_ context.Context, _ domain.RepoInfo) (string, error) {
	return m.path, m.err
}

type mockScanner struct {
	findings []domain.Finding
	err      error
}

func (m *mockScanner) Scan(_ context.Context, _ domain.RepoInfo, _ string) ([]domain.Finding, error) {
	return m.findings, m.err
}

type mockAnalyzer struct {
	report      domain.UtilityReport
	err         error
	callCount   int
}

func (m *mockAnalyzer) Analyze(_ context.Context, _ domain.RepoSummary) (domain.UtilityReport, error) {
	m.callCount++
	return m.report, m.err
}

type mockRepo struct {
	saved []domain.ScanResult
	err   error
}

func (m *mockRepo) Save(_ context.Context, result domain.ScanResult) error {
	m.saved = append(m.saved, result)
	return m.err
}

func (m *mockRepo) LoadAll(_ context.Context) ([]domain.ScanResult, error) {
	return m.saved, nil
}

type mockBroadcaster struct {
	events []domain.ProgressEvent
}

func (m *mockBroadcaster) Broadcast(event domain.ProgressEvent) {
	m.events = append(m.events, event)
}

func (m *mockBroadcaster) Register(_ chan domain.ProgressEvent)   {}
func (m *mockBroadcaster) Unregister(_ chan domain.ProgressEvent) {}

// --- Helpers ---

func twoRepos() []domain.RepoInfo {
	return []domain.RepoInfo{
		{Name: "repo-a", FullName: "user/repo-a"},
		{Name: "repo-b", FullName: "user/repo-b"},
	}
}

func oneFinding(repo string) []domain.Finding {
	return []domain.Finding{
		{
			ID:          "f1",
			Repo:        repo,
			FilePath:    "secrets.txt",
			Line:        42,
			PatternName: "aws-key",
			Severity:    domain.SeverityHigh,
		},
	}
}

// --- Factory helper ---

func staticFactory(p domain.RepoProvider) ProviderFactory {
	return func(_ string) (domain.RepoProvider, error) {
		return p, nil
	}
}

// --- Tests ---

func TestScanService_FlagOFF_Flow(t *testing.T) {
	ctx := context.Background()

	analyzer := &mockAnalyzer{}
	broadcaster := &mockBroadcaster{}
	repo := &mockRepo{}

	svc := NewScanService(
		staticFactory(&mockProvider{repos: twoRepos()}),
		&mockCloner{path: t.TempDir()},
		&mockScanner{findings: oneFinding("repo-a")},
		repo,
		broadcaster,
		analyzer,
		false, // AI disabled
	)

	if err := svc.Run(ctx, domain.ScanRequest{Username: "user", Provider: "github"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Analyzer must NOT have been called
	if analyzer.callCount != 0 {
		t.Errorf("expected analyzer not called, got %d calls", analyzer.callCount)
	}

	// Two repos saved
	if len(repo.saved) != 2 {
		t.Errorf("expected 2 saved results, got %d", len(repo.saved))
	}

	// No utility report on results
	for _, r := range repo.saved {
		if r.UtilityReport != nil {
			t.Errorf("expected no utility report when AI is disabled, got one for %s", r.Repo.Name)
		}
	}

	// Events broadcast — check scan_start and scan_complete present
	assertEventPresent(t, broadcaster.events, domain.EventScanStart)
	assertEventPresent(t, broadcaster.events, domain.EventScanComplete)
	assertEventPresent(t, broadcaster.events, domain.EventFindingFound)
	assertEventPresent(t, broadcaster.events, domain.EventRepoDone)
}

func TestScanService_FlagON_Flow(t *testing.T) {
	ctx := context.Background()

	analyzer := &mockAnalyzer{
		report: domain.UtilityReport{
			Repo:      "repo-a",
			Score:     8,
			Reason:    "active project",
			CreatedAt: time.Now(),
		},
	}
	broadcaster := &mockBroadcaster{}
	repo := &mockRepo{}

	svc := NewScanService(
		staticFactory(&mockProvider{repos: twoRepos()}),
		&mockCloner{path: t.TempDir()},
		&mockScanner{findings: oneFinding("repo-a")},
		repo,
		broadcaster,
		analyzer,
		true, // AI enabled
	)

	if err := svc.Run(ctx, domain.ScanRequest{Username: "user", Provider: "github"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Analyzer called once per repo
	if analyzer.callCount != 2 {
		t.Errorf("expected analyzer called 2 times (one per repo), got %d", analyzer.callCount)
	}

	// Each saved result must have utility report
	for _, r := range repo.saved {
		if r.UtilityReport == nil {
			t.Errorf("expected utility report on result for %s", r.Repo.Name)
		}
	}

	assertEventPresent(t, broadcaster.events, domain.EventAIAnalysisStart)
	assertEventPresent(t, broadcaster.events, domain.EventAIAnalysisDone)
}

func TestScanService_ClonerError(t *testing.T) {
	ctx := context.Background()

	broadcaster := &mockBroadcaster{}
	repo := &mockRepo{}

	svc := NewScanService(
		staticFactory(&mockProvider{repos: twoRepos()}),
		&mockCloner{err: errors.New("git clone failed")},
		&mockScanner{findings: oneFinding("repo-a")},
		repo,
		broadcaster,
		nil,
		false,
	)

	// Should not panic — scan continues with all repos
	if err := svc.Run(ctx, domain.ScanRequest{Username: "user", Provider: "github"}); err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}

	// Error event broadcast for each repo that failed to clone
	errorCount := countEvents(broadcaster.events, domain.EventError)
	if errorCount < 2 {
		t.Errorf("expected at least 2 error events (one per clone failure), got %d", errorCount)
	}

	// Nothing saved (clone failed before scan)
	if len(repo.saved) != 0 {
		t.Errorf("expected 0 saved results, got %d", len(repo.saved))
	}

	// scan_complete still broadcast
	assertEventPresent(t, broadcaster.events, domain.EventScanComplete)
}

func TestScanService_ScannerError(t *testing.T) {
	ctx := context.Background()

	broadcaster := &mockBroadcaster{}
	repo := &mockRepo{}

	svc := NewScanService(
		staticFactory(&mockProvider{repos: twoRepos()}),
		&mockCloner{path: t.TempDir()},
		&mockScanner{err: errors.New("scanner blew up")},
		repo,
		broadcaster,
		nil,
		false,
	)

	if err := svc.Run(ctx, domain.ScanRequest{Username: "user", Provider: "github"}); err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}

	// Error event broadcast
	errorCount := countEvents(broadcaster.events, domain.EventError)
	if errorCount < 2 {
		t.Errorf("expected at least 2 error events (one per scanner failure), got %d", errorCount)
	}

	// Partial results still saved (empty findings)
	if len(repo.saved) != 2 {
		t.Errorf("expected 2 saved results (with empty findings), got %d", len(repo.saved))
	}
}

func TestScanService_EventOrder(t *testing.T) {
	ctx := context.Background()

	broadcaster := &mockBroadcaster{}

	svc := NewScanService(
		staticFactory(&mockProvider{repos: []domain.RepoInfo{{Name: "repo-a", FullName: "user/repo-a"}}}),
		&mockCloner{path: t.TempDir()},
		&mockScanner{findings: oneFinding("repo-a")},
		&mockRepo{},
		broadcaster,
		nil,
		false,
	)

	if err := svc.Run(ctx, domain.ScanRequest{Username: "user", Provider: "github"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := broadcaster.events
	// Expected order: scan_start → repo_start → (file_scanned) → finding_found → repo_done → scan_complete
	orderedTypes := []domain.EventType{
		domain.EventScanStart,
		domain.EventRepoStart,
		domain.EventFindingFound,
		domain.EventRepoDone,
		domain.EventScanComplete,
	}

	lastIdx := -1
	for _, want := range orderedTypes {
		idx := findEventIndex(events, want, lastIdx+1)
		if idx == -1 {
			t.Errorf("event %q not found after index %d", want, lastIdx)
			continue
		}
		lastIdx = idx
	}
}

// --- Assertion helpers ---

func assertEventPresent(t *testing.T, events []domain.ProgressEvent, eventType domain.EventType) {
	t.Helper()
	for _, e := range events {
		if e.Type == eventType {
			return
		}
	}
	t.Errorf("expected event %q to be broadcast, but it was not", eventType)
}

func countEvents(events []domain.ProgressEvent, eventType domain.EventType) int {
	count := 0
	for _, e := range events {
		if e.Type == eventType {
			count++
		}
	}
	return count
}

func findEventIndex(events []domain.ProgressEvent, eventType domain.EventType, fromIdx int) int {
	for i := fromIdx; i < len(events); i++ {
		if events[i].Type == eventType {
			return i
		}
	}
	return -1
}
