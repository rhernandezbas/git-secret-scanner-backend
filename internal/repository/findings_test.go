package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

func makeResult(fullName string, scannedAt time.Time) domain.ScanResult {
	return domain.ScanResult{
		Repo: domain.RepoInfo{
			Name:     fullName,
			FullName: fullName,
		},
		Findings:  []domain.Finding{},
		ScannedAt: scannedAt,
		Duration:  "1s",
	}
}

func TestSave_CreatesFileIfNotExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.json")

	repo := NewJSONFindingsRepository(path)
	result := makeResult("org/repo", time.Now())

	if err := repo.Save(context.Background(), result); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to be created")
	}
}

func TestSave_AppendsResults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.json")
	repo := NewJSONFindingsRepository(path)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	r1 := makeResult("org/repo-a", now)
	r2 := makeResult("org/repo-b", now)

	if err := repo.Save(ctx, r1); err != nil {
		t.Fatalf("Save r1: %v", err)
	}
	if err := repo.Save(ctx, r2); err != nil {
		t.Fatalf("Save r2: %v", err)
	}

	results, err := repo.LoadAll(ctx)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestLoadAll_ReturnsEmptyIfFileMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.json")
	repo := NewJSONFindingsRepository(path)

	results, err := repo.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll returned error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty slice, got %d", len(results))
	}
}

func TestLoadAll_HandlesCorruptJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.json")

	if err := os.WriteFile(path, []byte("not valid json {{{{"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	repo := NewJSONFindingsRepository(path)

	// Must not panic and must return empty slice
	results, err := repo.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll returned error on corrupt file: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty slice on corrupt file, got %d", len(results))
	}
}

func TestSave_Deduplication(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.json")
	repo := NewJSONFindingsRepository(path)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	result := makeResult("org/repo", now)

	if err := repo.Save(ctx, result); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	if err := repo.Save(ctx, result); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	results, err := repo.LoadAll(ctx)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result after dedup, got %d", len(results))
	}
}

func TestSave_DifferentScannedAtNotDeduped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.json")
	repo := NewJSONFindingsRepository(path)
	ctx := context.Background()

	t1 := time.Now().UTC().Truncate(time.Second)
	t2 := t1.Add(time.Minute)

	r1 := makeResult("org/repo", t1)
	r2 := makeResult("org/repo", t2)

	if err := repo.Save(ctx, r1); err != nil {
		t.Fatalf("Save r1: %v", err)
	}
	if err := repo.Save(ctx, r2); err != nil {
		t.Fatalf("Save r2: %v", err)
	}

	results, err := repo.LoadAll(ctx)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results (different scannedAt), got %d", len(results))
	}
}

func TestLoadAll_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.json")
	repo := NewJSONFindingsRepository(path)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	want := makeResult("org/round-trip", now)
	want.Findings = []domain.Finding{
		{
			ID:          "f1",
			Repo:        "org/round-trip",
			Provider:    "github",
			FilePath:    "config/secrets.env",
			Line:        42,
			PatternName: "aws-key",
			Category:    "cloud",
			Severity:    domain.SeverityCritical,
			Match:       "AKIA...XXXX",
			ScannedAt:   now,
		},
	}

	if err := repo.Save(ctx, want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	results, err := repo.LoadAll(ctx)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	got := results[0]
	if got.Repo.FullName != want.Repo.FullName {
		t.Errorf("FullName mismatch: got %q want %q", got.Repo.FullName, want.Repo.FullName)
	}
	if len(got.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(got.Findings))
	}
	if got.Findings[0].ID != "f1" {
		t.Errorf("finding ID mismatch: got %q want %q", got.Findings[0].ID, "f1")
	}
}
