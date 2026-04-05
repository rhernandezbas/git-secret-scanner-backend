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

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

// AnthropicAnalyzer calls the Anthropic Messages API.
type AnthropicAnalyzer struct {
	apiKey  string
	model   string
	baseURL string // overridable for tests
	client  *http.Client
}

// NewAnthropicAnalyzer creates a new AnthropicAnalyzer.
func NewAnthropicAnalyzer(apiKey, model string) *AnthropicAnalyzer {
	return &AnthropicAnalyzer{
		apiKey:  apiKey,
		model:   model,
		baseURL: anthropicAPIURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func (a *AnthropicAnalyzer) Analyze(ctx context.Context, summary domain.RepoSummary) (domain.UtilityReport, error) {
	prompt, err := buildPrompt(summary)
	if err != nil {
		return domain.UtilityReport{}, err
	}

	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: 256,
		Messages:  []anthropicMessage{{Role: "user", Content: prompt}},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return domain.UtilityReport{}, fmt.Errorf("anthropic: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return domain.UtilityReport{}, fmt.Errorf("anthropic: create request: %w", err)
	}
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return domain.UtilityReport{}, fmt.Errorf("anthropic: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return domain.UtilityReport{}, fmt.Errorf("anthropic: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return domain.UtilityReport{}, fmt.Errorf("anthropic: decode response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return domain.UtilityReport{}, fmt.Errorf("anthropic: empty content in response")
	}

	text := extractJSON(apiResp.Content[0].Text)

	var ai aiResponse
	if err := json.Unmarshal([]byte(text), &ai); err != nil {
		return domain.UtilityReport{}, fmt.Errorf("anthropic: parse AI JSON: %w", err)
	}

	return domain.UtilityReport{
		Repo:      summary.Name,
		Score:     ai.Score,
		Reason:    ai.Reason,
		Model:     a.model,
		Provider:  "anthropic",
		CreatedAt: time.Now().UTC(),
	}, nil
}
