package languages

import (
	"reflect"
	"testing"
)

func TestExtractEnvVarsFromJava_StaticPatterns(t *testing.T) {
	tests := []struct {
		name     string
		matches  []map[string]string
		expected []EnvVarMatch
	}{
		{
			name: "System.getenv with string literal",
			matches: []map[string]string{
				{
					"obj":    "System",
					"method": "getenv",
					"key":    `"API_KEY"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "API_KEY", IsPartial: false},
			},
		},
		{
			name: "System.getenv().get with string literal",
			matches: []map[string]string{
				{
					"obj":     "System",
					"method1": "getenv",
					"method2": "get",
					"key":     `"DATABASE_URL"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "DATABASE_URL", IsPartial: false},
			},
		},
		{
			name: "single quotes",
			matches: []map[string]string{
				{
					"obj":    "System",
					"method": "getenv",
					"key":    `'SECRET_KEY'`,
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
					"obj":    "System",
					"method": "getenv",
					"key":    `"KEY1"`,
				},
				{
					"obj":     "System",
					"method1": "getenv",
					"method2": "get",
					"key":     `"KEY2"`,
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
			result := ExtractEnvVarsFromJavaWithPartial(tt.matches)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractEnvVarsFromJava_DynamicPatterns(t *testing.T) {
	tests := []struct {
		name     string
		matches  []map[string]string
		expected []EnvVarMatch
	}{
		{
			name: "binary expression with System.getenv",
			matches: []map[string]string{
				{
					"obj":       "System",
					"method":    "getenv",
					"full_expr": `"prefix_" + var`,
				},
			},
			expected: []EnvVarMatch{
				{Key: `"prefix_" + var`, IsPartial: true, FullExpr: `"prefix_" + var`},
			},
		},
		{
			name: "binary expression with System.getenv().get",
			matches: []map[string]string{
				{
					"obj":       "System",
					"method1":   "getenv",
					"method2":   "get",
					"full_expr": `var + "_suffix"`,
				},
			},
			expected: []EnvVarMatch{
				{Key: "var + \"_suffix\"", IsPartial: true, FullExpr: "var + \"_suffix\""},
			},
		},
		{
			name: "variable reference with System.getenv",
			matches: []map[string]string{
				{
					"obj":    "System",
					"method": "getenv",
					"var":    "envVar",
				},
			},
			expected: []EnvVarMatch{
				{Key: "envVar", IsPartial: true, IsVarRef: true},
			},
		},
		{
			name: "variable reference with System.getenv().get",
			matches: []map[string]string{
				{
					"obj":     "System",
					"method1": "getenv",
					"method2": "get",
					"var":     "key",
				},
			},
			expected: []EnvVarMatch{
				{Key: "key", IsPartial: true, IsVarRef: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEnvVarsFromJavaWithPartial(tt.matches)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractEnvVarsFromJava_InvalidPatterns(t *testing.T) {
	tests := []struct {
		name    string
		matches []map[string]string
	}{
		{
			name: "wrong object name",
			matches: []map[string]string{
				{
					"obj":    "Runtime",
					"method": "getenv",
					"key":    `"KEY"`,
				},
			},
		},
		{
			name: "wrong method name",
			matches: []map[string]string{
				{
					"obj":    "System",
					"method": "getProperty",
					"key":    `"KEY"`,
				},
			},
		},
		{
			name: "wrong chained method",
			matches: []map[string]string{
				{
					"obj":     "System",
					"method1": "getenv",
					"method2": "put",
					"key":     `"KEY"`,
				},
			},
		},
		{
			name: "empty key",
			matches: []map[string]string{
				{
					"obj":    "System",
					"method": "getenv",
					"key":    `""`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEnvVarsFromJavaWithPartial(tt.matches)
			if len(result) != 0 {
				t.Errorf("Expected no matches, got %v", result)
			}
		})
	}
}

func TestExtractEnvVarsFromJava_Deduplication(t *testing.T) {
	matches := []map[string]string{
		{
			"obj":    "System",
			"method": "getenv",
			"key":    `"DUPLICATE_KEY"`,
		},
		{
			"obj":     "System",
			"method1": "getenv",
			"method2": "get",
			"key":     `"DUPLICATE_KEY"`,
		},
	}

	result := ExtractEnvVarsFromJavaWithPartial(matches)
	if len(result) != 1 {
		t.Errorf("Expected 1 match after deduplication, got %d", len(result))
	}
	if result[0].Key != "DUPLICATE_KEY" {
		t.Errorf("Expected key 'DUPLICATE_KEY', got '%s'", result[0].Key)
	}
}

func TestExtractEnvVarsFromJava_BackwardCompatibility(t *testing.T) {
	matches := []map[string]string{
		{
			"obj":    "System",
			"method": "getenv",
			"key":    `"STATIC_KEY"`,
		},
		{
			"obj":       "System",
			"method":    "getenv",
			"full_expr": `"prefix_" + var`,
		},
	}

	result := ExtractEnvVarsFromJava(matches)
	expected := []string{"STATIC_KEY"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

