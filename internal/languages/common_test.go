package languages

import (
	"testing"
)

func TestGetLanguageInfo(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		expected *LanguageInfo
	}{
		{
			name: "javascript",
			lang: "javascript",
			expected: &LanguageInfo{
				Query:                JavaScriptQuery,
				Extractor:            nil,
				ExtractorWithPartial: ExtractEnvVarsFromJS,
			},
		},
		{
			name: "typescript",
			lang: "typescript",
			expected: &LanguageInfo{
				Query:                JavaScriptQuery,
				Extractor:            nil,
				ExtractorWithPartial: ExtractEnvVarsFromJS,
			},
		},
		{
			name: "go",
			lang: "go",
			expected: &LanguageInfo{
				Query:                GoQuery,
				Extractor:            ExtractEnvVarsFromGo,
				ExtractorWithPartial: ExtractEnvVarsFromGoWithPartial,
			},
		},
		{
			name: "python",
			lang: "python",
			expected: &LanguageInfo{
				Query:                PythonQuery,
				Extractor:            ExtractEnvVarsFromPython,
				ExtractorWithPartial: ExtractEnvVarsFromPythonWithPartial,
			},
		},
		{
			name: "rust",
			lang: "rust",
			expected: &LanguageInfo{
				Query:                RustQuery,
				Extractor:            ExtractEnvVarsFromRust,
				ExtractorWithPartial: ExtractEnvVarsFromRustWithPartial,
			},
		},
		{
			name: "java",
			lang: "java",
			expected: &LanguageInfo{
				Query:                JavaQuery,
				Extractor:            ExtractEnvVarsFromJava,
				ExtractorWithPartial: ExtractEnvVarsFromJavaWithPartial,
			},
		},
		{
			name:     "unknown language",
			lang:     "unknown",
			expected: nil,
		},
		{
			name:     "empty string",
			lang:     "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLanguageInfo(tt.lang)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil for unknown language, got %v", result)
				}
				return
			}
			if result == nil {
				t.Errorf("Expected LanguageInfo, got nil")
				return
			}
			if result.Query != tt.expected.Query {
				t.Errorf("Query mismatch: expected %s, got %s", tt.expected.Query[:50], result.Query[:50])
			}
			if (tt.expected.Extractor == nil) != (result.Extractor == nil) {
				t.Errorf("Extractor mismatch: expected %v, got %v", tt.expected.Extractor != nil, result.Extractor != nil)
			}
			if (tt.expected.ExtractorWithPartial == nil) != (result.ExtractorWithPartial == nil) {
				t.Errorf("ExtractorWithPartial mismatch: expected %v, got %v", tt.expected.ExtractorWithPartial != nil, result.ExtractorWithPartial != nil)
			}
		})
	}
}

