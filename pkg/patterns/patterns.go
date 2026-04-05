package patterns

import "regexp"

type Category string
type Severity string

const (
	CategoryLLM           Category = "llm"
	CategoryCloud         Category = "cloud"
	CategorySourceControl Category = "source_control"
	CategoryPayment       Category = "payment"
	CategoryAuth          Category = "auth"
	CategoryDatabase      Category = "database"
	CategoryCommunication Category = "communication"
)

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
)

type Pattern struct {
	Name     string
	Category Category
	Severity Severity
	Regex    *regexp.Regexp
}

var All []Pattern

func init() {
	raw := []struct {
		Name     string
		Category Category
		Severity Severity
		Pattern  string
	}{
		// === LLM PROVIDERS ===
		{"openai_api_key", CategoryLLM, SeverityCritical, `sk-[a-zA-Z0-9]{48}`},
		{"openai_project_key", CategoryLLM, SeverityCritical, `sk-proj-[a-zA-Z0-9_\-]{40,}`},
		{"anthropic_api_key", CategoryLLM, SeverityCritical, `sk-ant-api03-[a-zA-Z0-9\-_]{93}`},
		{"anthropic_key_generic", CategoryLLM, SeverityCritical, `sk-ant-[a-zA-Z0-9\-_]{20,}`},
		{"huggingface_token", CategoryLLM, SeverityHigh, `hf_[a-zA-Z0-9]{37}`},
		{"replicate_token", CategoryLLM, SeverityHigh, `r8_[a-zA-Z0-9]{40}`},
		{"groq_api_key", CategoryLLM, SeverityCritical, `gsk_[a-zA-Z0-9]{52}`},
		{"google_ai_key", CategoryLLM, SeverityHigh, `AIza[0-9A-Za-z\-_]{35}`},
		{"cohere_api_key", CategoryLLM, SeverityHigh, `[a-zA-Z0-9]{40}(?:cohere|co-[a-zA-Z0-9])`},
		{"mistral_api_key", CategoryLLM, SeverityHigh, `[a-zA-Z0-9]{32}(?:mistral)?`},
		{"together_api_key", CategoryLLM, SeverityHigh, `[a-f0-9]{64}`},
		{"stability_api_key", CategoryLLM, SeverityHigh, `sk-[a-zA-Z0-9]{48}`},
		{"perplexity_api_key", CategoryLLM, SeverityHigh, `pplx-[a-zA-Z0-9]{48}`},
		{"ai21_api_key", CategoryLLM, SeverityHigh, `[a-zA-Z0-9]{32}`},
		{"deepinfra_api_key", CategoryLLM, SeverityHigh, `[a-zA-Z0-9_\-]{40,}`},
		{"openrouter_api_key", CategoryLLM, SeverityCritical, `sk-or-v1-[a-zA-Z0-9]{64}`},
		{"fireworks_api_key", CategoryLLM, SeverityHigh, `fw_[a-zA-Z0-9]{40}`},
		{"anyscale_api_key", CategoryLLM, SeverityHigh, `esecret_[a-zA-Z0-9]{40}`},
		{"voyage_api_key", CategoryLLM, SeverityHigh, `pa-[a-zA-Z0-9]{40}`},
		{"deepseek_api_key", CategoryLLM, SeverityHigh, `sk-[a-f0-9]{32}`},

		// === CLOUD ===
		{"aws_access_key", CategoryCloud, SeverityCritical, `AKIA[0-9A-Z]{16}`},
		{"aws_secret_key", CategoryCloud, SeverityCritical, `(?:aws_secret|AWS_SECRET)[_\s]*[=:][_\s]*["\']?([0-9a-zA-Z/+]{40})`},
		{"gcp_api_key", CategoryCloud, SeverityHigh, `AIza[0-9A-Za-z\-_]{35}`},
		{"azure_subscription_key", CategoryCloud, SeverityHigh, `[a-f0-9]{32}`},
		{"cloudflare_api_token", CategoryCloud, SeverityHigh, `[a-zA-Z0-9_\-]{40}`},
		{"digitalocean_token", CategoryCloud, SeverityHigh, `dop_v1_[a-f0-9]{64}`},
		{"heroku_api_key", CategoryCloud, SeverityHigh, `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`},

		// === SOURCE CONTROL & CI/CD ===
		{"github_pat", CategorySourceControl, SeverityCritical, `ghp_[a-zA-Z0-9]{36}`},
		{"github_oauth", CategorySourceControl, SeverityCritical, `gho_[a-zA-Z0-9]{36}`},
		{"github_app_token", CategorySourceControl, SeverityCritical, `ghs_[a-zA-Z0-9]{36}`},
		{"github_refresh_token", CategorySourceControl, SeverityHigh, `ghr_[a-zA-Z0-9]{36}`},
		{"gitlab_token", CategorySourceControl, SeverityCritical, `glpat-[a-zA-Z0-9\-_]{20}`},
		{"circleci_token", CategorySourceControl, SeverityHigh, `[0-9a-f]{40}`},
		{"travis_token", CategorySourceControl, SeverityHigh, `travis_?token["\s]*[=:]["\s]*[a-zA-Z0-9_\-]{20,}`},

		// === PAYMENT ===
		{"stripe_secret_key", CategoryPayment, SeverityCritical, `sk_live_[0-9a-zA-Z]{24}`},
		{"stripe_restricted_key", CategoryPayment, SeverityCritical, `rk_live_[0-9a-zA-Z]{24}`},

		// === COMMUNICATION ===
		{"twilio_account_sid", CategoryCommunication, SeverityHigh, `AC[a-zA-Z0-9]{32}`},
		{"twilio_auth_token", CategoryCommunication, SeverityCritical, `SK[0-9a-fA-F]{32}`},
		{"sendgrid_api_key", CategoryCommunication, SeverityCritical, `SG\.[a-zA-Z0-9_\-]{22}\.[a-zA-Z0-9_\-]{43}`},
		{"mailgun_api_key", CategoryCommunication, SeverityHigh, `key-[a-zA-Z0-9]{32}`},

		// === AUTH & CRYPTO ===
		{"jwt_token", CategoryAuth, SeverityHigh, `eyJ[a-zA-Z0-9_\-]+\.[a-zA-Z0-9_\-]+\.[a-zA-Z0-9_\-]+`},
		{"rsa_private_key", CategoryAuth, SeverityCritical, `-----BEGIN RSA PRIVATE KEY-----`},
		{"ec_private_key", CategoryAuth, SeverityCritical, `-----BEGIN EC PRIVATE KEY-----`},
		{"private_key_generic", CategoryAuth, SeverityCritical, `-----BEGIN PRIVATE KEY-----`},
		{"pgp_private_key", CategoryAuth, SeverityCritical, `-----BEGIN PGP PRIVATE KEY BLOCK-----`},

		// === DATABASE ===
		{"postgres_uri", CategoryDatabase, SeverityCritical, `postgres(?:ql)?://[^:]+:[^@]+@[^\s"']+`},
		{"mysql_uri", CategoryDatabase, SeverityCritical, `mysql://[^:]+:[^@]+@[^\s"']+`},
		{"mongodb_uri", CategoryDatabase, SeverityCritical, `mongodb(?:\+srv)?://[^:]+:[^@]+@[^\s"']+`},
		{"redis_uri", CategoryDatabase, SeverityHigh, `redis://[^:]*:[^@]+@[^\s"']+`},
	}

	for _, r := range raw {
		compiled, err := regexp.Compile(r.Pattern)
		if err != nil {
			panic("invalid pattern " + r.Name + ": " + err.Error())
		}
		All = append(All, Pattern{
			Name:     r.Name,
			Category: r.Category,
			Severity: r.Severity,
			Regex:    compiled,
		})
	}
}

// Redact returns first4...last4 of a matched string.
func Redact(match string) string {
	if len(match) <= 8 {
		return "****"
	}
	return match[:4] + "..." + match[len(match)-4:]
}
