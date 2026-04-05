package analyzer

import (
	"fmt"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

// New creates a new Analyzer for the given provider.
// model can be empty to use the provider default.
func New(provider, apiKey, model string) (domain.Analyzer, error) {
	switch provider {
	case "anthropic":
		m := model
		if m == "" {
			m = "claude-haiku-4-5-20251001"
		}
		return NewAnthropicAnalyzer(apiKey, m), nil
	case "openai":
		m := model
		if m == "" {
			m = "gpt-4o-mini"
		}
		return NewOpenAIAnalyzer(apiKey, m), nil
	case "groq":
		m := model
		if m == "" {
			m = "llama-3.1-8b-instant"
		}
		return NewGroqAnalyzer(apiKey, m), nil
	default:
		return nil, fmt.Errorf("unknown AI provider: %s (use anthropic, openai, or groq)", provider)
	}
}
