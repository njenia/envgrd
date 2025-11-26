package languages

import (
	"reflect"
	"testing"
)

func TestExtractEnvVarsFromRust_StaticPatterns(t *testing.T) {
	tests := []struct {
		name     string
		matches  []map[string]string
		expected []EnvVarMatch
	}{
		{
			name: "env::var with string literal",
			matches: []map[string]string{
				{
					"path": "env",
					"fn":   "var",
					"key":  `"API_KEY"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "API_KEY", IsPartial: false},
			},
		},
		{
			name: "std::env::var with string literal",
			matches: []map[string]string{
				{
					"path1": "std",
					"path2": "env",
					"fn":    "var",
					"key":   `"DATABASE_URL"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "DATABASE_URL", IsPartial: false},
			},
		},
		{
			name: "env::var_os with string literal",
			matches: []map[string]string{
				{
					"path": "env",
					"fn":   "var_os",
					"key":  `"SECRET_KEY"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "SECRET_KEY", IsPartial: false},
			},
		},
		{
			name: "std::env::var_os with string literal",
			matches: []map[string]string{
				{
					"path1": "std",
					"path2": "env",
					"fn":    "var_os",
					"key":   `"CONFIG_KEY"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "CONFIG_KEY", IsPartial: false},
			},
		},
		{
			name: "multiple static patterns",
			matches: []map[string]string{
				{
					"path": "env",
					"fn":   "var",
					"key":  `"KEY1"`,
				},
				{
					"path1": "std",
					"path2": "env",
					"fn":    "var",
					"key":   `"KEY2"`,
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
			result := ExtractEnvVarsFromRustWithPartial(tt.matches)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractEnvVarsFromRust_DynamicPatterns(t *testing.T) {
	tests := []struct {
		name     string
		matches  []map[string]string
		expected []EnvVarMatch
	}{
		{
			name: "binary expression with env::var",
			matches: []map[string]string{
				{
					"path":     "env",
					"fn":       "var",
					"full_expr": `"prefix_" + var`,
				},
			},
			expected: []EnvVarMatch{
				{Key: `"prefix_" + var`, IsPartial: true, FullExpr: `"prefix_" + var`},
			},
		},
		{
			name: "binary expression with std::env::var",
			matches: []map[string]string{
				{
					"path1":    "std",
					"path2":    "env",
					"fn":       "var",
					"full_expr": `var + "_suffix"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "var + \"_suffix\"", IsPartial: true, FullExpr: "var + \"_suffix\""},
			},
		},
		{
			name: "variable reference with env::var",
			matches: []map[string]string{
				{
					"path": "env",
					"fn":   "var",
					"var":  "envVar",
				},
			},
			expected: []EnvVarMatch{
				{Key: "envVar", IsPartial: true, IsVarRef: true},
			},
		},
		{
			name: "variable reference with std::env::var",
			matches: []map[string]string{
				{
					"path1": "std",
					"path2": "env",
					"fn":    "var",
					"var":   "key",
				},
			},
			expected: []EnvVarMatch{
				{Key: "key", IsPartial: true, IsVarRef: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEnvVarsFromRustWithPartial(tt.matches)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractEnvVarsFromRust_InvalidPatterns(t *testing.T) {
	tests := []struct {
		name    string
		matches []map[string]string
	}{
		{
			name: "wrong function name",
			matches: []map[string]string{
				{
					"path": "env",
					"fn":   "args",
					"key":  `"KEY"`,
				},
			},
		},
		{
			name: "wrong path",
			matches: []map[string]string{
				{
					"path": "std",
					"fn":   "var",
					"key":  `"KEY"`,
				},
			},
		},
		{
			name: "wrong std path",
			matches: []map[string]string{
				{
					"path1": "core",
					"path2": "env",
					"fn":    "var",
					"key":   `"KEY"`,
				},
			},
		},
		{
			name: "empty key",
			matches: []map[string]string{
				{
					"path": "env",
					"fn":   "var",
					"key":  `""`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEnvVarsFromRustWithPartial(tt.matches)
			if len(result) != 0 {
				t.Errorf("Expected no matches, got %v", result)
			}
		})
	}
}

func TestExtractEnvVarsFromRust_Deduplication(t *testing.T) {
	matches := []map[string]string{
		{
			"path": "env",
			"fn":   "var",
			"key":  `"DUPLICATE_KEY"`,
		},
		{
			"path1": "std",
			"path2": "env",
			"fn":    "var",
			"key":   `"DUPLICATE_KEY"`,
		},
	}

	result := ExtractEnvVarsFromRustWithPartial(matches)
	if len(result) != 1 {
		t.Errorf("Expected 1 match after deduplication, got %d", len(result))
	}
	if result[0].Key != "DUPLICATE_KEY" {
		t.Errorf("Expected key 'DUPLICATE_KEY', got '%s'", result[0].Key)
	}
}

func TestExtractEnvVarsFromRust_BackwardCompatibility(t *testing.T) {
	matches := []map[string]string{
		{
			"path": "env",
			"fn":   "var",
			"key":  `"STATIC_KEY"`,
		},
		{
			"path":     "env",
			"fn":       "var",
			"full_expr": `"prefix_" + var`,
		},
	}

	result := ExtractEnvVarsFromRust(matches)
	expected := []string{"STATIC_KEY"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

