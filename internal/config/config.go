package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	GithubToken       string
	GitlabToken       string
	AIAnalysisEnabled bool
	AIProvider        string // anthropic | openai | groq
	AIAPIKey          string
	AIModel           string
	MaxRepoSizeMB     int
	FindingsFile      string
	TempDir           string
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Port:         getEnv("PORT", "8080"),
		GithubToken:  getEnv("GITHUB_TOKEN", ""),
		GitlabToken:  getEnv("GITLAB_TOKEN", ""),
		AIProvider:   getEnv("AI_PROVIDER", "anthropic"),
		AIModel:      getEnv("AI_MODEL", ""),
		FindingsFile: getEnv("FINDINGS_FILE", "findings.json"),
		TempDir:      getEnv("TEMP_DIR", "/tmp"),
	}

	var err error
	cfg.AIAnalysisEnabled, err = strconv.ParseBool(getEnv("AI_ANALYSIS_ENABLED", "false"))
	if err != nil {
		return nil, fmt.Errorf("invalid AI_ANALYSIS_ENABLED value: %w", err)
	}

	cfg.MaxRepoSizeMB, err = strconv.Atoi(getEnv("MAX_REPO_SIZE_MB", "500"))
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_REPO_SIZE_MB value: %w", err)
	}

	cfg.AIAPIKey = os.Getenv("AI_API_KEY")
	if cfg.AIAnalysisEnabled && cfg.AIAPIKey == "" {
		return nil, fmt.Errorf("AI_API_KEY is required when AI_ANALYSIS_ENABLED=true")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
