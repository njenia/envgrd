package languages

import (
	"reflect"
	"testing"
)

func TestExtractEnvVarsFromJS_StaticPatterns(t *testing.T) {
	tests := []struct {
		name     string
		matches  []map[string]string
		expected []EnvVarMatch
	}{
		{
			name: "dot notation",
			matches: []map[string]string{
				{
					"obj":  "process",
					"prop": "env",
					"key":  "API_KEY",
				},
			},
			expected: []EnvVarMatch{
				{Key: "API_KEY", IsPartial: false},
			},
		},
		{
			name: "bracket notation with double quotes",
			matches: []map[string]string{
				{
					"obj":  "process",
					"prop": "env",
					"key":  `"DATABASE_URL"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "DATABASE_URL", IsPartial: false},
			},
		},
		{
			name: "bracket notation with single quotes",
			matches: []map[string]string{
				{
					"obj":  "process",
					"prop": "env",
					"key":  `'SECRET_KEY'`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "SECRET_KEY", IsPartial: false},
			},
		},
		{
			name: "multiple static patterns",
			matches: []map[string]string{
				{
					"obj":  "process",
					"prop": "env",
					"key":  "KEY1",
				},
				{
					"obj":  "process",
					"prop": "env",
					"key":  `"KEY2"`,
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
			result := ExtractEnvVarsFromJS(tt.matches)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractEnvVarsFromJS_DynamicPatterns(t *testing.T) {
	tests := []struct {
		name     string
		matches  []map[string]string
		expected []EnvVarMatch
	}{
		{
			name: "binary expression with prefix",
			matches: []map[string]string{
				{
					"obj":      "process",
					"prop":     "env",
					"full_expr": `"prefix_" + var`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "prefix_", IsPartial: true, FullExpr: "prefix_\" + var"},
			},
		},
		{
			name: "binary expression with suffix",
			matches: []map[string]string{
				{
					"obj":      "process",
					"prop":     "env",
					"full_expr": `var + "_suffix"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "_suffix", IsPartial: true, FullExpr: "var + \"_suffix\""},
			},
		},
		{
			name: "variable reference",
			matches: []map[string]string{
				{
					"obj":  "process",
					"prop": "env",
					"var":  "envVar",
				},
			},
			expected: []EnvVarMatch{
				{Key: "envVar", IsPartial: true, IsVarRef: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEnvVarsFromJS(tt.matches)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d matches, got %d", len(tt.expected), len(result))
			}
			for i, exp := range tt.expected {
				if i >= len(result) {
					t.Errorf("Missing expected match: %v", exp)
					continue
				}
				if result[i].Key != exp.Key || result[i].IsPartial != exp.IsPartial || result[i].IsVarRef != exp.IsVarRef {
					t.Errorf("Match %d: Expected %v, got %v", i, exp, result[i])
				}
			}
		})
	}
}

func TestExtractEnvVarsFromJS_InvalidPatterns(t *testing.T) {
	tests := []struct {
		name    string
		matches []map[string]string
	}{
		{
			name: "wrong object name",
			matches: []map[string]string{
				{
					"obj":  "window",
					"prop": "env",
					"key":  "KEY",
				},
			},
		},
		{
			name: "wrong property name",
			matches: []map[string]string{
				{
					"obj":  "process",
					"prop": "argv",
					"key":  "KEY",
				},
			},
		},
		{
			name: "empty key",
			matches: []map[string]string{
				{
					"obj":  "process",
					"prop": "env",
					"key":  "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEnvVarsFromJS(tt.matches)
			if len(result) != 0 {
				t.Errorf("Expected no matches, got %v", result)
			}
		})
	}
}

func TestExtractEnvVarsFromJS_Deduplication(t *testing.T) {
	matches := []map[string]string{
		{
			"obj":  "process",
			"prop": "env",
			"key":  "DUPLICATE_KEY",
		},
		{
			"obj":  "process",
			"prop": "env",
			"key":  `"DUPLICATE_KEY"`,
		},
		{
			"obj":  "process",
			"prop": "env",
			"key":  "DUPLICATE_KEY",
		},
	}

	result := ExtractEnvVarsFromJS(matches)
	if len(result) != 1 {
		t.Errorf("Expected 1 match after deduplication, got %d", len(result))
	}
	if result[0].Key != "DUPLICATE_KEY" {
		t.Errorf("Expected key 'DUPLICATE_KEY', got '%s'", result[0].Key)
	}
}

func TestExtractFirstString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"double quotes", `"prefix_" + var`, "prefix_"},
		{"single quotes", `'prefix_' + var`, "prefix_"},
		{"backticks", "`prefix_` + var", "prefix_"},
		{"no quotes", "var + other", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFirstString(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestExtractLastString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"double quotes", `var + "_suffix"`, "_suffix"},
		{"single quotes", `var + '_suffix'`, "_suffix"},
		{"backticks", "var + `_suffix`", "_suffix"},
		{"no quotes", "var + other", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLastString(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

