package analyzer

// EnvUsage represents a single usage of an environment variable in code
type EnvUsage struct {
	Key          string // The environment variable key
	File         string // File path where it's used
	Line         int    // Line number where it's used
	CodeSnippet  string // Code snippet from the line where it's used
	InIgnoredPath bool  // True if this usage is in a folder that should be ignored
	IsPartial    bool   // True if this is a partial match from dynamic code (e.g., "prefix_" + var)
	IsVarRef     bool   // True if this is a variable reference pattern (e.g., process.env[a])
	FullExpr     string // Full expression for dynamic patterns (e.g., "prefix_" + var)
}

// EnvFile represents a parsed environment file
type EnvFile struct {
	Path string            // Path to the env file
	Vars map[string]string // Key-value pairs from the file
}

// ScanResult contains the complete analysis results
type ScanResult struct {
	CodeKeys           []EnvUsage            // All env var usages found in code
	EnvKeys            map[string]string     // All env vars from .env files
	Missing            map[string][]EnvUsage // Missing keys (in code but not in .env) grouped by key
	PartialMatches     map[string][]EnvUsage // Partial matches (dynamic code patterns) grouped by prefix/suffix
	Unused             []string              // Unused keys (in .env but not in code)
	IgnoredMissing     int                   // Count of missing variables that were ignored via config
	IgnoredFromFolders int                   // Count of unique variables found in ignored folders
}

