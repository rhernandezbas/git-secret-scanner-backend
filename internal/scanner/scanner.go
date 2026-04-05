package scanner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
	"github.com/rhernandezba/git-secret-scanner/backend/pkg/patterns"
)

const maxFileSizeBytes = 1 * 1024 * 1024 // 1MB

var skipDirs = map[string]bool{
	".git": true, "vendor": true, "node_modules": true,
	"dist": true, "build": true, ".next": true,
}

// FileScanner scans a local directory for secret patterns.
type FileScanner struct{}

// New returns a new FileScanner.
func New() *FileScanner {
	return &FileScanner{}
}

// Scan walks localPath recursively and returns all findings.
func (s *FileScanner) Scan(ctx context.Context, repo domain.RepoInfo, localPath string) ([]domain.Finding, error) {
	var findings []domain.Finding
	counter := 0

	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Size() > maxFileSizeBytes {
			return nil
		}

		found, err := scanFile(ctx, path, localPath, repo, &counter)
		if err != nil {
			return nil // skip files with errors
		}
		findings = append(findings, found...)
		return nil
	})

	return findings, err
}

func scanFile(ctx context.Context, path, basePath string, repo domain.RepoInfo, counter *int) ([]domain.Finding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Check for binary content in first 512 bytes.
	header := make([]byte, 512)
	n, _ := f.Read(header)
	if !utf8.Valid(header[:n]) {
		return nil, nil // binary file — skip
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}

	relPath, _ := filepath.Rel(basePath, path)
	var findings []domain.Finding
	lineNum := 0
	sc := bufio.NewScanner(f)

	for sc.Scan() {
		if ctx.Err() != nil {
			return findings, ctx.Err()
		}
		lineNum++
		line := sc.Text()

		for _, p := range patterns.All {
			match := p.Regex.FindString(line)
			if match == "" {
				continue
			}
			*counter++
			findings = append(findings, domain.Finding{
				ID:          fmt.Sprintf("%d-%d", time.Now().UnixNano(), *counter),
				Repo:        repo.Name,
				Provider:    repo.Provider,
				FilePath:    relPath,
				Line:        lineNum,
				PatternName: p.Name,
				Category:    string(p.Category),
				Severity:    domain.Severity(p.Severity),
				Match:       patterns.Redact(match),
				ScannedAt:   time.Now(),
			})
		}
	}

	return findings, sc.Err()
}
