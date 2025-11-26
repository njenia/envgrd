package languages

import (
	"reflect"
	"testing"
)

func TestExtractEnvVarsFromPython_StaticPatterns(t *testing.T) {
	tests := []struct {
		name     string
		matches  []map[string]string
		expected []EnvVarMatch
	}{
		{
			name: "os.environ with string literal",
			matches: []map[string]string{
				{
					"obj":  "os",
					"attr": "environ",
					"key":  `"API_KEY"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "API_KEY", IsPartial: false},
			},
		},
		{
			name: "os.getenv with string literal",
			matches: []map[string]string{
				{
					"obj2": "os",
					"fn":   "getenv",
					"key":  `"DATABASE_URL"`,
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
					"obj":  "os",
					"attr": "environ",
					"key":  `"KEY1"`,
				},
				{
					"obj2": "os",
					"fn":   "getenv",
					"key":  `"KEY2"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "KEY1", IsPartial: false},
				{Key: "KEY2", IsPartial: false},
			},
		},
		{
			name: "single quotes",
			matches: []map[string]string{
				{
					"obj":  "os",
					"attr": "environ",
					"key":  `'SECRET_KEY'`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "SECRET_KEY", IsPartial: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEnvVarsFromPythonWithPartial(tt.matches)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractEnvVarsFromPython_DynamicPatterns(t *testing.T) {
	tests := []struct {
		name     string
		matches  []map[string]string
		expected []EnvVarMatch
	}{
		{
			name: "os.environ with binary expression",
			matches: []map[string]string{
				{
					"obj":       "os",
					"attr":      "environ",
					"full_expr": `"prefix_" + var`,
				},
			},
			expected: []EnvVarMatch{
				{Key: `"prefix_" + var`, IsPartial: true, FullExpr: `"prefix_" + var`},
			},
		},
		{
			name: "os.getenv with binary expression",
			matches: []map[string]string{
				{
					"obj2":      "os",
					"fn":        "getenv",
					"full_expr": `var + "_suffix"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "var + \"_suffix\"", IsPartial: true, FullExpr: "var + \"_suffix\""},
			},
		},
		{
			name: "variable reference in os.environ",
			matches: []map[string]string{
				{
					"obj":  "os",
					"attr": "environ",
					"var":  "envVar",
				},
			},
			expected: []EnvVarMatch{
				{Key: "envVar", IsPartial: true, IsVarRef: true},
			},
		},
		{
			name: "variable reference in os.getenv",
			matches: []map[string]string{
				{
					"obj2": "os",
					"fn":   "getenv",
					"var":  "key",
				},
			},
			expected: []EnvVarMatch{
				{Key: "key", IsPartial: true, IsVarRef: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEnvVarsFromPythonWithPartial(tt.matches)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractEnvVarsFromPython_InvalidPatterns(t *testing.T) {
	tests := []struct {
		name    string
		matches []map[string]string
	}{
		{
			name: "wrong object name",
			matches: []map[string]string{
				{
					"obj":  "sys",
					"attr": "environ",
					"key":  `"KEY"`,
				},
			},
		},
		{
			name: "wrong attribute name",
			matches: []map[string]string{
				{
					"obj":  "os",
					"attr": "path",
					"key":  `"KEY"`,
				},
			},
		},
		{
			name: "wrong function name",
			matches: []map[string]string{
				{
					"obj2": "os",
					"fn":   "getcwd",
					"key":  `"KEY"`,
				},
			},
		},
		{
			name: "empty key",
			matches: []map[string]string{
				{
					"obj":  "os",
					"attr": "environ",
					"key":  `""`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEnvVarsFromPythonWithPartial(tt.matches)
			if len(result) != 0 {
				t.Errorf("Expected no matches, got %v", result)
			}
		})
	}
}

func TestExtractEnvVarsFromPython_Deduplication(t *testing.T) {
	matches := []map[string]string{
		{
			"obj":  "os",
			"attr": "environ",
			"key":  `"DUPLICATE_KEY"`,
		},
		{
			"obj":  "os",
			"attr": "environ",
			"key":  `"DUPLICATE_KEY"`,
		},
		{
			"obj2": "os",
			"fn":   "getenv",
			"key":  `"DUPLICATE_KEY"`,
		},
	}

	result := ExtractEnvVarsFromPythonWithPartial(matches)
	if len(result) != 1 {
		t.Errorf("Expected 1 match after deduplication, got %d", len(result))
	}
	if result[0].Key != "DUPLICATE_KEY" {
		t.Errorf("Expected key 'DUPLICATE_KEY', got '%s'", result[0].Key)
	}
}

func TestExtractEnvVarsFromPython_BackwardCompatibility(t *testing.T) {
	matches := []map[string]string{
		{
			"obj":  "os",
			"attr": "environ",
			"key":  `"STATIC_KEY"`,
		},
		{
			"obj":       "os",
			"attr":      "environ",
			"full_expr": `"prefix_" + var`,
		},
	}

	result := ExtractEnvVarsFromPython(matches)
	expected := []string{"STATIC_KEY"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
