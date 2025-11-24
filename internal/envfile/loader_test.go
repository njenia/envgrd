package envfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvFile(t *testing.T) {
	// Create a temporary .env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	content := `# This is a comment
KEY1=value1
KEY2=value2
KEY3="quoted value"
KEY4='single quoted'

# Empty line above
KEY5=value5
`
	err := os.WriteFile(envPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	vars, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("Failed to parse .env file: %v", err)
	}

	expected := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
		"KEY3": "quoted value",
		"KEY4": "single quoted",
		"KEY5": "value5",
	}

	if len(vars) != len(expected) {
		t.Errorf("Expected %d vars, got %d", len(expected), len(vars))
	}

	for key, expectedValue := range expected {
		if actualValue, ok := vars[key]; !ok {
			t.Errorf("Missing key: %s", key)
		} else if actualValue != expectedValue {
			t.Errorf("Key %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}
}

func TestParseEnvFile_NonExistent(t *testing.T) {
	vars, err := parseEnvFile("/nonexistent/.env")
	if err != nil {
		t.Errorf("Non-existent file should return empty map, not error: %v", err)
	}
	if len(vars) != 0 {
		t.Errorf("Expected empty map for non-existent file, got %d vars", len(vars))
	}
}

func TestLoader_Load(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .env file
	env1 := filepath.Join(tmpDir, ".env")
	os.WriteFile(env1, []byte("KEY1=value1\nKEY2=value2\n"), 0644)

	// Create .env.local file (should override .env)
	env2 := filepath.Join(tmpDir, ".env.local")
	os.WriteFile(env2, []byte("KEY2=overridden\nKEY3=value3\n"), 0644)

	loader := NewLoader()
	vars, err := loader.Load(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load env files: %v", err)
	}

	// KEY1 should be from .env
	if vars["KEY1"] != "value1" {
		t.Errorf("KEY1: expected value1, got %s", vars["KEY1"])
	}

	// KEY2 should be overridden by .env.local
	if vars["KEY2"] != "overridden" {
		t.Errorf("KEY2: expected overridden, got %s", vars["KEY2"])
	}

	// KEY3 should be from .env.local
	if vars["KEY3"] != "value3" {
		t.Errorf("KEY3: expected value3, got %s", vars["KEY3"])
	}
}

