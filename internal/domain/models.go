package domain

import "time"

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

type Finding struct {
	ID          string    `json:"id"`
	Repo        string    `json:"repo"`
	Provider    string    `json:"provider"`
	FilePath    string    `json:"file_path"`
	Line        int       `json:"line"`
	PatternName string    `json:"pattern_name"`
	Category    string    `json:"category"`
	Severity    Severity  `json:"severity"`
	Match       string    `json:"match"` // redacted: first4...last4
	ScannedAt   time.Time `json:"scanned_at"`
}

type UtilityReport struct {
	Repo      string    `json:"repo"`
	Score     int       `json:"score"` // 0-10
	Reason    string    `json:"reason"`
	Model     string    `json:"model"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
}

type RepoInfo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	CloneURL    string `json:"clone_url"`
	SizeKB      int    `json:"size_kb"`
	Language    string `json:"language"`
	Stars       int    `json:"stars"`
	Provider    string `json:"provider"` // github | gitlab
}

type RepoSummary struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Languages     []string `json:"languages"`
	FileCount     int      `json:"file_count"`
	SizeKB        int      `json:"size_kb"`
	TopFiles      []string `json:"top_files"`
	ReadmeExcerpt string   `json:"readme_excerpt"` // first 500 chars
}

type ScanResult struct {
	Repo          RepoInfo       `json:"repo"`
	Findings      []Finding      `json:"findings"`
	UtilityReport *UtilityReport `json:"utility_report,omitempty"`
	Duration      string         `json:"duration"`
	ScannedAt     time.Time      `json:"scanned_at"`
	Error         string         `json:"error,omitempty"`
}

type EventType string

const (
	EventScanStart       EventType = "scan_start"
	EventRepoStart       EventType = "repo_start"
	EventFileScanned     EventType = "file_scanned"
	EventFindingFound    EventType = "finding_found"
	EventAIAnalysisStart EventType = "ai_analysis_start"
	EventAIAnalysisDone  EventType = "ai_analysis_done"
	EventRepoDone        EventType = "repo_done"
	EventScanComplete    EventType = "scan_complete"
	EventError           EventType = "error"
)

type ProgressEvent struct {
	Type    EventType   `json:"type"`
	Repo    string      `json:"repo,omitempty"`
	Message string      `json:"message"`
	Payload interface{} `json:"payload,omitempty"`
}

type ScanRequest struct {
	Username string `json:"username"`
	Provider string `json:"provider"` // github | gitlab
}
