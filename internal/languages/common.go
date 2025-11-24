package languages

// EnvVarMatch represents a matched environment variable (static or partial)
type EnvVarMatch struct {
	Key          string
	IsPartial    bool
	IsVarRef     bool   // True if this is a variable reference (e.g., process.env[a])
	FullExpr     string // Full expression for dynamic patterns (e.g., "prefix_" + var)
}

// LanguageInfo contains query and extraction function for a language
type LanguageInfo struct {
	Query     string
	Extractor func([]map[string]string) []string // Returns []string for backward compatibility
	// For JavaScript/TypeScript, we'll use a special handler
	ExtractorWithPartial func([]map[string]string) []EnvVarMatch // Returns matches with partial info
}

// GetLanguageInfo returns the query and extractor for a given language
func GetLanguageInfo(lang string) *LanguageInfo {
	switch lang {
	case "javascript", "typescript":
		return &LanguageInfo{
			Query:                JavaScriptQuery,
			Extractor:            nil, // Not used for JS/TS
			ExtractorWithPartial: ExtractEnvVarsFromJS,
		}
	case "go":
		return &LanguageInfo{
			Query:                GoQuery,
			Extractor:            ExtractEnvVarsFromGo, // For backward compatibility
			ExtractorWithPartial: ExtractEnvVarsFromGoWithPartial,
		}
	case "python":
		return &LanguageInfo{
			Query:                PythonQuery,
			Extractor:            ExtractEnvVarsFromPython, // For backward compatibility
			ExtractorWithPartial: ExtractEnvVarsFromPythonWithPartial,
		}
	case "rust":
		return &LanguageInfo{
			Query:                RustQuery,
			Extractor:            ExtractEnvVarsFromRust, // For backward compatibility
			ExtractorWithPartial: ExtractEnvVarsFromRustWithPartial,
		}
	case "java":
		return &LanguageInfo{
			Query:                JavaQuery,
			Extractor:            ExtractEnvVarsFromJava, // For backward compatibility
			ExtractorWithPartial: ExtractEnvVarsFromJavaWithPartial,
		}
	default:
		return nil
	}
}

