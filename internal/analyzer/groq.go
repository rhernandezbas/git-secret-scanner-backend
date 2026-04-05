package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

const groqAPIURL = "https://api.groq.com/openai/v1/chat/completions"

// GroqAnalyzer calls the Groq API (OpenAI-compatible).
type GroqAnalyzer struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewGroqAnalyzer creates a new GroqAnalyzer.
func NewGroqAnalyzer(apiKey, model string) *GroqAnalyzer {
	return &GroqAnalyzer{
		apiKey:  apiKey,
		model:   model,
		baseURL: groqAPIURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *GroqAnalyzer) Analyze(ctx context.Context, summary domain.RepoSummary) (domain.UtilityReport, error) {
	prompt, err := buildPrompt(summary)
	if err != nil {
		return domain.UtilityReport{}, err
	}

	reqBody := openaiRequest{
		Model:     g.model,
		MaxTokens: 256,
		Messages:  []openaiMessage{{Role: "user", Content: prompt}},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return domain.UtilityReport{}, fmt.Errorf("groq: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return domain.UtilityReport{}, fmt.Errorf("groq: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("content-type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return domain.UtilityReport{}, fmt.Errorf("groq: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return domain.UtilityReport{}, fmt.Errorf("groq: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return domain.UtilityReport{}, fmt.Errorf("groq: decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return domain.UtilityReport{}, fmt.Errorf("groq: empty choices in response")
	}

	text := extractJSON(apiResp.Choices[0].Message.Content)

	var ai aiResponse
	if err := json.Unmarshal([]byte(text), &ai); err != nil {
		return domain.UtilityReport{}, fmt.Errorf("groq: parse AI JSON: %w", err)
	}

	return domain.UtilityReport{
		Repo:      summary.Name,
		Score:     ai.Score,
		Reason:    ai.Reason,
		Model:     g.model,
		Provider:  "groq",
		CreatedAt: time.Now().UTC(),
	}, nil
}
