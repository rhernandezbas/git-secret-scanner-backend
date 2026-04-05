package patterns

import (
	"testing"
)

func TestRedact(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"short input", "abc", "****"},
		{"exactly 8 chars", "abcdefgh", "****"},
		{"long input", "sk-abcdefghij1234", "sk-a...1234"},
		{"very long", "AKIAIOSFODNN7EXAMPLE", "AKIA...MPLE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Redact(tt.input)
			if got != tt.want {
				t.Errorf("Redact(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAllPatternsCompile(t *testing.T) {
	if len(All) == 0 {
		t.Fatal("no patterns loaded")
	}
	t.Logf("total patterns: %d", len(All))
}

func TestPatternPositives(t *testing.T) {
	tests := []struct {
		patternName string
		sample      string
	}{
		{"openai_api_key", "sk-" + repeat("A", 48)},
		{"openai_project_key", "sk-proj-" + repeat("A", 40)},
		{"anthropic_api_key", "sk-ant-api03-" + repeat("A", 93)},
		{"anthropic_key_generic", "sk-ant-" + repeat("A", 20)},
		{"huggingface_token", "hf_" + repeat("a", 37)},
		{"replicate_token", "r8_" + repeat("a", 40)},
		{"groq_api_key", "gsk_" + repeat("a", 52)},
		{"google_ai_key", "AIza" + repeat("a", 35)},
		{"perplexity_api_key", "pplx-" + repeat("a", 48)},
		{"openrouter_api_key", "sk-or-v1-" + repeat("a", 64)},
		{"fireworks_api_key", "fw_" + repeat("a", 40)},
		{"anyscale_api_key", "esecret_" + repeat("a", 40)},
		{"voyage_api_key", "pa-" + repeat("a", 40)},
		{"aws_access_key", "AKIAIOSFODNN7EXAMPLE"},
		{"digitalocean_token", "dop_v1_" + repeat("a", 64)},
		{"heroku_api_key", "12345678-1234-1234-1234-123456789abc"},
		{"github_pat", "ghp_" + repeat("a", 36)},
		{"github_oauth", "gho_" + repeat("a", 36)},
		{"github_app_token", "ghs_" + repeat("a", 36)},
		{"github_refresh_token", "ghr_" + repeat("a", 36)},
		{"gitlab_token", "glpat-" + repeat("a", 20)},
		{"stripe_secret_key", "sk_live_" + repeat("a", 24)},
		{"stripe_restricted_key", "rk_live_" + repeat("a", 24)},
		{"twilio_account_sid", "AC" + repeat("a", 32)},
		{"twilio_auth_token", "SK" + repeat("0", 32)},
		{"sendgrid_api_key", "SG." + repeat("a", 22) + "." + repeat("b", 43)},
		{"mailgun_api_key", "key-" + repeat("a", 32)},
		{"jwt_token", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ1c2VyIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
		{"rsa_private_key", "-----BEGIN RSA PRIVATE KEY-----"},
		{"ec_private_key", "-----BEGIN EC PRIVATE KEY-----"},
		{"private_key_generic", "-----BEGIN PRIVATE KEY-----"},
		{"pgp_private_key", "-----BEGIN PGP PRIVATE KEY BLOCK-----"},
		{"postgres_uri", "postgresql://user:password@localhost:5432/db"},
		{"mysql_uri", "mysql://user:password@localhost:3306/db"},
		{"mongodb_uri", "mongodb://user:password@localhost:27017/db"},
		{"redis_uri", "redis://:password@localhost:6379"},
	}

	byName := make(map[string]Pattern, len(All))
	for _, p := range All {
		byName[p.Name] = p
	}

	for _, tt := range tests {
		t.Run("positive/"+tt.patternName, func(t *testing.T) {
			p, ok := byName[tt.patternName]
			if !ok {
				t.Fatalf("pattern %q not found", tt.patternName)
			}
			if !p.Regex.MatchString(tt.sample) {
				t.Errorf("pattern %q did not match sample %q", tt.patternName, tt.sample)
			}
		})
	}
}

func TestPatternNegatives(t *testing.T) {
	// Patterns with specific prefixes should NOT match clean/random text
	prefixedPatterns := []struct {
		patternName string
		clean       string
	}{
		{"openai_api_key", "this is a clean line with no secrets"},
		{"github_pat", "just a normal comment"},
		{"aws_access_key", "some random BKIAIOSFODNN7EXAMPLE text"},
		{"stripe_secret_key", "sk_test_notlive123456789012345"},
		{"digitalocean_token", "dop_v2_short"},
		{"sendgrid_api_key", "SG.tooshort"},
		{"mailgun_api_key", "key-tooshort"},
	}

	byName := make(map[string]Pattern, len(All))
	for _, p := range All {
		byName[p.Name] = p
	}

	for _, tt := range prefixedPatterns {
		t.Run("negative/"+tt.patternName, func(t *testing.T) {
			p, ok := byName[tt.patternName]
			if !ok {
				t.Fatalf("pattern %q not found", tt.patternName)
			}
			if p.Regex.MatchString(tt.clean) {
				t.Errorf("pattern %q unexpectedly matched clean string %q", tt.patternName, tt.clean)
			}
		})
	}
}

func repeat(s string, n int) string {
	result := make([]byte, n*len(s))
	for i := range result {
		result[i] = s[i%len(s)]
	}
	return string(result)
}
