package analyzer

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

const promptTmpl = `Analyze the following GitHub/GitLab repository and determine if it is useful or interesting.
Consider: active development, code quality indicators, documentation, unique functionality, stars.

Repository: {{.Name}}
Description: {{.Description}}
Languages: {{.Languages}}
Files: {{.FileCount}}
Size: {{.SizeKB}} KB
Top files: {{.TopFiles}}
README excerpt: {{.ReadmeExcerpt}}

Respond with ONLY valid JSON in this exact format:
{"score": <0-10>, "reason": "<one sentence explanation>"}`

var tmpl = template.Must(template.New("prompt").Parse(promptTmpl))

func buildPrompt(summary domain.RepoSummary) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, summary); err != nil {
		return "", fmt.Errorf("building prompt: %w", err)
	}
	return buf.String(), nil
}

// aiResponse is the expected JSON from the AI.
type aiResponse struct {
	Score  int    `json:"score"`
	Reason string `json:"reason"`
}

// extractJSON strips optional markdown code fences and whitespace from s.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	// Strip ```json ... ``` or ``` ... ```
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	return s
}
