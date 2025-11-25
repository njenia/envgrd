package analyzer

import (
	"testing"

	"github.com/jenian/envgrd/internal/config"
)

func TestAnalyze_MissingKeys(t *testing.T) {
	codeUsages := []EnvUsage{
		{Key: "STRIPE_KEY", File: "payments.js", Line: 10},
		{Key: "DATABASE_URL", File: "db.go", Line: 20},
		{Key: "API_KEY", File: "api.js", Line: 30},
	}

	envVars := map[string]string{
		"API_KEY": "test123",
	}

	cfg := &config.Config{}
	envKeySources := make(map[string]string)
	result := Analyze(codeUsages, envVars, envVars, envKeySources, cfg)

	// Should find 2 missing keys
	if len(result.Missing) != 2 {
		t.Errorf("Expected 2 missing keys, got %d", len(result.Missing))
	}

	if _, ok := result.Missing["STRIPE_KEY"]; !ok {
		t.Error("STRIPE_KEY should be missing")
	}

	if _, ok := result.Missing["DATABASE_URL"]; !ok {
		t.Error("DATABASE_URL should be missing")
	}

	if _, ok := result.Missing["API_KEY"]; ok {
		t.Error("API_KEY should not be missing")
	}
}

func TestAnalyze_UnusedKeys(t *testing.T) {
	codeUsages := []EnvUsage{
		{Key: "STRIPE_KEY", File: "payments.js", Line: 10},
	}

	envVars := map[string]string{
		"STRIPE_KEY": "sk_test_123",
		"OLD_API_KEY": "old123",
		"UNUSED_VAR": "unused",
	}

	cfg := &config.Config{}
	envKeySources := make(map[string]string)
	result := Analyze(codeUsages, envVars, envVars, envKeySources, cfg)

	// Should find 2 unused keys
	if len(result.Unused) != 2 {
		t.Errorf("Expected 2 unused keys, got %d", len(result.Unused))
	}

	unusedMap := make(map[string]bool)
	for _, key := range result.Unused {
		unusedMap[key] = true
	}

	if !unusedMap["OLD_API_KEY"] {
		t.Error("OLD_API_KEY should be unused")
	}

	if !unusedMap["UNUSED_VAR"] {
		t.Error("UNUSED_VAR should be unused")
	}

	if unusedMap["STRIPE_KEY"] {
		t.Error("STRIPE_KEY should not be unused")
	}
}

func TestAnalyze_NoIssues(t *testing.T) {
	codeUsages := []EnvUsage{
		{Key: "STRIPE_KEY", File: "payments.js", Line: 10},
		{Key: "DATABASE_URL", File: "db.go", Line: 20},
	}

	envVars := map[string]string{
		"STRIPE_KEY":  "sk_test_123",
		"DATABASE_URL": "postgres://localhost/db",
	}

	cfg := &config.Config{}
	envKeySources := make(map[string]string)
	result := Analyze(codeUsages, envVars, envVars, envKeySources, cfg)

	if len(result.Missing) != 0 {
		t.Errorf("Expected no missing keys, got %d", len(result.Missing))
	}

	if len(result.Unused) != 0 {
		t.Errorf("Expected no unused keys, got %d", len(result.Unused))
	}
}

func TestAnalyze_IgnoredMissing(t *testing.T) {
	codeUsages := []EnvUsage{
		{Key: "STRIPE_KEY", File: "payments.js", Line: 10},
		{Key: "DATABASE_URL", File: "db.go", Line: 20},
		{Key: "CUSTOM_VAR", File: "custom.go", Line: 5},
	}

	envVars := map[string]string{
		"STRIPE_KEY": "sk_test_123",
	}

	cfg := &config.Config{
		Ignores: config.IgnoresConfig{
			Missing: []string{"CUSTOM_VAR"},
		},
	}
	envKeySources := make(map[string]string)

	result := Analyze(codeUsages, envVars, envVars, envKeySources, cfg)

	// Should find 1 missing key (DATABASE_URL), CUSTOM_VAR should be ignored
	if len(result.Missing) != 1 {
		t.Errorf("Expected 1 missing key, got %d", len(result.Missing))
	}

	if _, ok := result.Missing["DATABASE_URL"]; !ok {
		t.Error("DATABASE_URL should be missing")
	}

	if _, ok := result.Missing["CUSTOM_VAR"]; ok {
		t.Error("CUSTOM_VAR should be ignored, not reported as missing")
	}

	if result.IgnoredMissing != 1 {
		t.Errorf("Expected 1 ignored missing variable, got %d", result.IgnoredMissing)
	}
}

