package languages

import (
	"reflect"
	"testing"
)

func TestExtractEnvVarsFromGo_StaticPatterns(t *testing.T) {
	tests := []struct {
		name     string
		matches  []map[string]string
		expected []EnvVarMatch
	}{
		{
			name: "os.Getenv with string literal",
			matches: []map[string]string{
				{
					"obj": "os",
					"fn":  "Getenv",
					"key": `"API_KEY"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "API_KEY", IsPartial: false},
			},
		},
		{
			name: "backticks",
			matches: []map[string]string{
				{
					"obj": "os",
					"fn":  "Getenv",
					"key": "`DATABASE_URL`",
				},
			},
			expected: []EnvVarMatch{
				{Key: "DATABASE_URL", IsPartial: false},
			},
		},
		{
			name: "multiple static patterns",
			matches: []map[string]string{
				{
					"obj": "os",
					"fn":  "Getenv",
					"key": `"KEY1"`,
				},
				{
					"obj": "os",
					"fn":  "Getenv",
					"key": `"KEY2"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "KEY1", IsPartial: false},
				{Key: "KEY2", IsPartial: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEnvVarsFromGoWithPartial(tt.matches)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractEnvVarsFromGo_DynamicPatterns(t *testing.T) {
	tests := []struct {
		name     string
		matches  []map[string]string
		expected []EnvVarMatch
	}{
		{
			name: "binary expression",
			matches: []map[string]string{
				{
					"obj":       "os",
					"fn":        "Getenv",
					"full_expr": `"prefix_" + var`,
				},
			},
			expected: []EnvVarMatch{
				{Key: `"prefix_" + var`, IsPartial: true, FullExpr: `"prefix_" + var`},
			},
		},
		{
			name: "variable reference",
			matches: []map[string]string{
				{
					"obj": "os",
					"fn":  "Getenv",
					"var": "envVar",
				},
			},
			expected: []EnvVarMatch{
				{Key: "envVar", IsPartial: true, IsVarRef: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEnvVarsFromGoWithPartial(tt.matches)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractEnvVarsFromGo_InvalidPatterns(t *testing.T) {
	tests := []struct {
		name    string
		matches []map[string]string
	}{
		{
			name: "wrong object name",
			matches: []map[string]string{
				{
					"obj": "fmt",
					"fn":  "Getenv",
					"key": `"KEY"`,
				},
			},
		},
		{
			name: "wrong function name",
			matches: []map[string]string{
				{
					"obj": "os",
					"fn":  "Getpid",
					"key": `"KEY"`,
				},
			},
		},
		{
			name: "empty key",
			matches: []map[string]string{
				{
					"obj": "os",
					"fn":  "Getenv",
					"key": `""`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEnvVarsFromGoWithPartial(tt.matches)
			if len(result) != 0 {
				t.Errorf("Expected no matches, got %v", result)
			}
		})
	}
}

func TestExtractEnvVarsFromGo_Deduplication(t *testing.T) {
	matches := []map[string]string{
		{
			"obj": "os",
			"fn":  "Getenv",
			"key": `"DUPLICATE_KEY"`,
		},
		{
			"obj": "os",
			"fn":  "Getenv",
			"key": `"DUPLICATE_KEY"`,
		},
	}

	result := ExtractEnvVarsFromGoWithPartial(matches)
	if len(result) != 1 {
		t.Errorf("Expected 1 match after deduplication, got %d", len(result))
	}
	if result[0].Key != "DUPLICATE_KEY" {
		t.Errorf("Expected key 'DUPLICATE_KEY', got '%s'", result[0].Key)
	}
}

func TestExtractEnvVarsFromGo_BackwardCompatibility(t *testing.T) {
	matches := []map[string]string{
		{
			"obj": "os",
			"fn":  "Getenv",
			"key": `"STATIC_KEY"`,
		},
		{
			"obj":       "os",
			"fn":        "Getenv",
			"full_expr": `"prefix_" + var`,
		},
	}

	result := ExtractEnvVarsFromGo(matches)
	expected := []string{"STATIC_KEY"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"double quotes", `"test"`, "test"},
		{"single quotes", `'test'`, "test"},
		{"backticks", "`test`", "test"},
		{"no quotes", "test", "test"},
		{"empty", "", ""},
		{"only opening quote", `"test`, `"test`},
		{"only closing quote", `test"`, `test"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimQuotes(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

