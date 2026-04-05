package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/handler"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/service"
)

// mockRepoProvider implements domain.RepoProvider
type mockRepoProvider struct{}

func (m *mockRepoProvider) ListPublicRepos(_ context.Context, _ string) ([]domain.RepoInfo, error) {
	return []domain.RepoInfo{}, nil
}

// mockCloner implements domain.Cloner
type mockCloner struct{}

func (m *mockCloner) Clone(_ context.Context, _ domain.RepoInfo) (string, error) {
	return "/tmp/test-repo", nil
}

// mockScanner implements domain.Scanner
type mockScanner struct{}

func (m *mockScanner) Scan(_ context.Context, _ domain.RepoInfo, _ string) ([]domain.Finding, error) {
	return []domain.Finding{}, nil
}

// mockFindingsRepo implements domain.FindingsRepository
type mockFindingsRepo struct {
	results []domain.ScanResult
}

func (m *mockFindingsRepo) Save(_ context.Context, result domain.ScanResult) error {
	m.results = append(m.results, result)
	return nil
}

func (m *mockFindingsRepo) LoadAll(_ context.Context) ([]domain.ScanResult, error) {
	return m.results, nil
}

// mockBroadcaster implements domain.EventBroadcaster
type mockBroadcaster struct{}

func (m *mockBroadcaster) Broadcast(_ domain.ProgressEvent)          {}
func (m *mockBroadcaster) Register(_ chan domain.ProgressEvent)       {}
func (m *mockBroadcaster) Unregister(_ chan domain.ProgressEvent)     {}

func newTestScanService() *service.ScanService {
	return service.NewScanService(
		func(_ string) (domain.RepoProvider, error) { return &mockRepoProvider{}, nil },
		&mockCloner{},
		&mockScanner{},
		&mockFindingsRepo{},
		&mockBroadcaster{},
		nil,   // analyzer
		false, // aiEnabled
	)
}

func TestHealth(t *testing.T) {
	svc := newTestScanService()
	h := handler.NewScanHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	h.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

func TestScan_ValidBody_Returns202(t *testing.T) {
	svc := newTestScanService()
	h := handler.NewScanHandler(svc)

	payload := `{"username":"testuser","provider":"github"}`
	req := httptest.NewRequest(http.MethodPost, "/scan", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Scan(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "scan started" {
		t.Errorf("expected status=scan started, got %q", body["status"])
	}
}

func TestScan_InvalidJSON_Returns400(t *testing.T) {
	svc := newTestScanService()
	h := handler.NewScanHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/scan", strings.NewReader("not-json"))
	rr := httptest.NewRecorder()

	h.Scan(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestScan_MissingUsername_Returns400(t *testing.T) {
	svc := newTestScanService()
	h := handler.NewScanHandler(svc)

	payload := `{"provider":"github"}`
	req := httptest.NewRequest(http.MethodPost, "/scan", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Scan(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestScan_DefaultsProviderToGithub(t *testing.T) {
	svc := newTestScanService()
	h := handler.NewScanHandler(svc)

	payload := `{"username":"testuser"}`
	req := httptest.NewRequest(http.MethodPost, "/scan", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Scan(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["provider"] != "github" {
		t.Errorf("expected provider=github, got %q", body["provider"])
	}
}

func TestScan_ViaHttptestServer(t *testing.T) {
	svc := newTestScanService()
	h := handler.NewScanHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/scan", h.Scan)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Give background goroutine time to complete
	time.Sleep(10 * time.Millisecond)

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
