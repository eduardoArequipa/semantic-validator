package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("TYPESAFE_API_KEY", "test-token")
	t.Setenv("SEMANTIC_VALIDATOR_API_KEY", "local-dev")
	t.Setenv("PORT", "")
	t.Setenv("TYPESAFE_BASE_URL", "")
	t.Setenv("JEV_MODEL", "")
	t.Setenv("JEV_TIMEOUT", "")
	t.Setenv("TRUST_PROXY_HEADERS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 || cfg.JevModel != "jev-latest" || cfg.JevTimeout.String() != "30s" || cfg.TrustProxyHeaders {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadTrustsProxyHeadersWhenEnabled(t *testing.T) {
	t.Setenv("TYPESAFE_API_KEY", "test-token")
	t.Setenv("SEMANTIC_VALIDATOR_API_KEY", "local-dev")
	t.Setenv("TRUST_PROXY_HEADERS", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.TrustProxyHeaders {
		t.Fatal("expected proxy headers to be trusted")
	}
}

func TestLoadRequiresAPIKey(t *testing.T) {
	t.Setenv("TYPESAFE_API_KEY", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing API key error")
	}
}

func TestLoadRequiresSemanticValidatorKey(t *testing.T) {
	t.Setenv("API_KEYS_FILE", "")
	t.Setenv("TYPESAFE_API_KEY", "test-token")
	t.Setenv("SEMANTIC_VALIDATOR_API_KEY", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing Semantic Validator API key error")
	}
}

func TestLoadIndividualKeysWithoutLegacyKey(t *testing.T) {
	t.Setenv("TYPESAFE_API_KEY", "test-token")
	t.Setenv("SEMANTIC_VALIDATOR_API_KEY", "")
	t.Setenv("API_KEYS_FILE", "data/keys.json")
	cfg, err := Load()
	if err != nil || cfg.APIKeysFile != "data/keys.json" {
		t.Fatalf("individual keys config: %v", err)
	}
}

func TestLoadDocsOnlyNeedsNoKeys(t *testing.T) {
	t.Setenv("DOCS_ONLY", "true")
	t.Setenv("TYPESAFE_API_KEY", "")
	t.Setenv("SEMANTIC_VALIDATOR_API_KEY", "")
	t.Setenv("API_KEYS_FILE", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.DocsOnly || cfg.Port != 8080 || cfg.JevAPIKey != "" {
		t.Fatalf("unexpected docs-only config: %+v", cfg)
	}
}

func TestLoadRejectsInvalidDocsOnly(t *testing.T) {
	t.Setenv("DOCS_ONLY", "maybe")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid DOCS_ONLY error")
	}
}
