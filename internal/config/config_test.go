package config

import (
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear all relevant env vars to test defaults
	vars := []string{
		"PORT", "GITHUB_TOKEN", "GITLAB_TOKEN",
		"AI_ANALYSIS_ENABLED", "AI_PROVIDER", "AI_API_KEY", "AI_MODEL",
		"MAX_REPO_SIZE_MB", "FINDINGS_FILE", "TEMP_DIR",
	}
	for _, v := range vars {
		t.Setenv(v, "")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port: want 8080, got %s", cfg.Port)
	}
	if cfg.AIProvider != "anthropic" {
		t.Errorf("AIProvider: want anthropic, got %s", cfg.AIProvider)
	}
	if cfg.AIAnalysisEnabled != false {
		t.Errorf("AIAnalysisEnabled: want false, got %v", cfg.AIAnalysisEnabled)
	}
	if cfg.MaxRepoSizeMB != 500 {
		t.Errorf("MaxRepoSizeMB: want 500, got %d", cfg.MaxRepoSizeMB)
	}
	if cfg.FindingsFile != "findings.json" {
		t.Errorf("FindingsFile: want findings.json, got %s", cfg.FindingsFile)
	}
	if cfg.TempDir != "/tmp" {
		t.Errorf("TempDir: want /tmp, got %s", cfg.TempDir)
	}
}

func TestLoad_EnvVarsOverrideDefaults(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("GITHUB_TOKEN", "gh-token")
	t.Setenv("GITLAB_TOKEN", "gl-token")
	t.Setenv("AI_ANALYSIS_ENABLED", "true")
	t.Setenv("AI_PROVIDER", "openai")
	t.Setenv("AI_API_KEY", "sk-test")
	t.Setenv("AI_MODEL", "gpt-4")
	t.Setenv("MAX_REPO_SIZE_MB", "1000")
	t.Setenv("FINDINGS_FILE", "output.json")
	t.Setenv("TEMP_DIR", "/var/tmp")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port: want 9090, got %s", cfg.Port)
	}
	if cfg.GithubToken != "gh-token" {
		t.Errorf("GithubToken: want gh-token, got %s", cfg.GithubToken)
	}
	if cfg.GitlabToken != "gl-token" {
		t.Errorf("GitlabToken: want gl-token, got %s", cfg.GitlabToken)
	}
	if cfg.AIAnalysisEnabled != true {
		t.Errorf("AIAnalysisEnabled: want true, got %v", cfg.AIAnalysisEnabled)
	}
	if cfg.AIProvider != "openai" {
		t.Errorf("AIProvider: want openai, got %s", cfg.AIProvider)
	}
	if cfg.AIAPIKey != "sk-test" {
		t.Errorf("AIAPIKey: want sk-test, got %s", cfg.AIAPIKey)
	}
	if cfg.AIModel != "gpt-4" {
		t.Errorf("AIModel: want gpt-4, got %s", cfg.AIModel)
	}
	if cfg.MaxRepoSizeMB != 1000 {
		t.Errorf("MaxRepoSizeMB: want 1000, got %d", cfg.MaxRepoSizeMB)
	}
	if cfg.FindingsFile != "output.json" {
		t.Errorf("FindingsFile: want output.json, got %s", cfg.FindingsFile)
	}
	if cfg.TempDir != "/var/tmp" {
		t.Errorf("TempDir: want /var/tmp, got %s", cfg.TempDir)
	}
}

func TestLoad_RequiredVars(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name: "AI_API_KEY required when AI_ANALYSIS_ENABLED=true",
			envVars: map[string]string{
				"AI_ANALYSIS_ENABLED": "true",
				"AI_API_KEY":          "",
			},
			wantErr: true,
			errMsg:  "AI_API_KEY is required when AI_ANALYSIS_ENABLED=true",
		},
		{
			name: "AI_API_KEY not required when AI_ANALYSIS_ENABLED=false",
			envVars: map[string]string{
				"AI_ANALYSIS_ENABLED": "false",
				"AI_API_KEY":          "",
			},
			wantErr: false,
		},
		{
			name: "invalid AI_ANALYSIS_ENABLED value",
			envVars: map[string]string{
				"AI_ANALYSIS_ENABLED": "notabool",
			},
			wantErr: true,
			errMsg:  "invalid AI_ANALYSIS_ENABLED value",
		},
		{
			name: "invalid MAX_REPO_SIZE_MB value",
			envVars: map[string]string{
				"MAX_REPO_SIZE_MB": "notanumber",
			},
			wantErr: true,
			errMsg:  "invalid MAX_REPO_SIZE_MB value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset all relevant env vars
			allVars := []string{
				"PORT", "GITHUB_TOKEN", "GITLAB_TOKEN",
				"AI_ANALYSIS_ENABLED", "AI_PROVIDER", "AI_API_KEY", "AI_MODEL",
				"MAX_REPO_SIZE_MB", "FINDINGS_FILE", "TEMP_DIR",
			}
			for _, v := range allVars {
				t.Setenv(v, "")
			}
			// Apply test-specific env vars
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			_, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if tt.errMsg != "" {
					if err.Error() == "" {
						t.Errorf("expected error containing %q, got empty error", tt.errMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}
