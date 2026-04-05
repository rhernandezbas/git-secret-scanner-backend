package repository

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

type JSONFindingsRepository struct {
	filePath string
	mu       sync.Mutex
}

func NewJSONFindingsRepository(filePath string) *JSONFindingsRepository {
	return &JSONFindingsRepository{filePath: filePath}
}

func (r *JSONFindingsRepository) Save(ctx context.Context, result domain.ScanResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, _ := r.loadRaw()

	// Dedup by repo full name + scanned_at
	for _, e := range existing {
		if e.Repo.FullName == result.Repo.FullName && e.ScannedAt.Equal(result.ScannedAt) {
			return nil // already saved
		}
	}

	existing = append(existing, result)
	return r.writeRaw(existing)
}

func (r *JSONFindingsRepository) LoadAll(ctx context.Context) ([]domain.ScanResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	results, _ := r.loadRaw()
	return results, nil
}

func (r *JSONFindingsRepository) loadRaw() ([]domain.ScanResult, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return []domain.ScanResult{}, nil
	}
	var results []domain.ScanResult
	if err := json.Unmarshal(data, &results); err != nil {
		return []domain.ScanResult{}, nil // corrupt file: return empty
	}
	return results, nil
}

func (r *JSONFindingsRepository) writeRaw(results []domain.ScanResult) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.filePath, data, 0644)
}
