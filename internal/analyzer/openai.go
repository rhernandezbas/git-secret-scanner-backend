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

const openaiAPIURL = "https://api.openai.com/v1/chat/completions"

// OpenAIAnalyzer calls the OpenAI Chat Completions API.
type OpenAIAnalyzer struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewOpenAIAnalyzer creates a new OpenAIAnalyzer.
func NewOpenAIAnalyzer(apiKey, model string) *OpenAIAnalyzer {
	return &OpenAIAnalyzer{
		apiKey:  apiKey,
		model:   model,
		baseURL: openaiAPIURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type openaiRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []openaiMessage `json:"messages"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (o *OpenAIAnalyzer) Analyze(ctx context.Context, summary domain.RepoSummary) (domain.UtilityReport, error) {
	prompt, err := buildPrompt(summary)
	if err != nil {
		return domain.UtilityReport{}, err
	}

	reqBody := openaiRequest{
		Model:     o.model,
		MaxTokens: 256,
		Messages:  []openaiMessage{{Role: "user", Content: prompt}},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return domain.UtilityReport{}, fmt.Errorf("openai: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return domain.UtilityReport{}, fmt.Errorf("openai: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("content-type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return domain.UtilityReport{}, fmt.Errorf("openai: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return domain.UtilityReport{}, fmt.Errorf("openai: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return domain.UtilityReport{}, fmt.Errorf("openai: decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return domain.UtilityReport{}, fmt.Errorf("openai: empty choices in response")
	}

	text := extractJSON(apiResp.Choices[0].Message.Content)

	var ai aiResponse
	if err := json.Unmarshal([]byte(text), &ai); err != nil {
		return domain.UtilityReport{}, fmt.Errorf("openai: parse AI JSON: %w", err)
	}

	return domain.UtilityReport{
		Repo:      summary.Name,
		Score:     ai.Score,
		Reason:    ai.Reason,
		Model:     o.model,
		Provider:  "openai",
		CreatedAt: time.Now().UTC(),
	}, nil
}
