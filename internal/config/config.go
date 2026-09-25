package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DocsOnly             bool
	APIKeysFile          string
	Port                 int
	SemanticValidatorKey string
	TrustProxyHeaders    bool
	JevBaseURL           string
	JevAPIKey            string
	JevModel             string
	JevTimeout           time.Duration
}

func Load() (Config, error) {
	docsOnly, err := boolEnv("DOCS_ONLY", false)
	if err != nil {
		return Config{}, err
	}

	trustProxyHeaders, err := boolEnv("TRUST_PROXY_HEADERS", false)
	if err != nil {
		return Config{}, err
	}

	port, err := intEnv("PORT", 8080)
	if err != nil {
		return Config{}, err
	}
	if port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be between 1 and 65535")
	}
	if docsOnly {
		return Config{DocsOnly: true, Port: port}, nil
	}

	timeout, err := durationEnv("JEV_TIMEOUT", 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	if timeout <= 0 {
		return Config{}, fmt.Errorf("JEV_TIMEOUT must be positive")
	}

	apiKey := os.Getenv("TYPESAFE_API_KEY")
	if apiKey == "" {
		return Config{}, fmt.Errorf("TYPESAFE_API_KEY is required")
	}
	semanticValidatorKey := os.Getenv("SEMANTIC_VALIDATOR_API_KEY")
	if semanticValidatorKey == "" && os.Getenv("API_KEYS_FILE") == "" {
		return Config{}, fmt.Errorf("SEMANTIC_VALIDATOR_API_KEY is required")
	}

	return Config{
		DocsOnly:             false,
		APIKeysFile:          os.Getenv("API_KEYS_FILE"),
		Port:                 port,
		SemanticValidatorKey: semanticValidatorKey,
		TrustProxyHeaders:    trustProxyHeaders,
		JevBaseURL:           envOrDefault("TYPESAFE_BASE_URL", "https://api.typesafe.ai"),
		JevAPIKey:            apiKey,
		JevModel:             envOrDefault("JEV_MODEL", "jev-latest"),
		JevTimeout:           timeout,
	}, nil
}

func boolEnv(name string, fallback bool) (bool, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", name, err)
	}
	return parsed, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func intEnv(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	return parsed, nil
}

func durationEnv(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", name, err)
	}
	return parsed, nil
}
