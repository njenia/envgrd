package analyzer

import (
	"strings"

	"github.com/jenian/envgrd/internal/config"
)

// Analyze compares code-discovered environment variables with those in .env files
// envVars: all environment variables (from .env files + exported env vars) - used for missing check
// envVarsFromFiles: only variables from .env files - used for unused check
// envKeySources: maps variable key to source file path
// cfg: configuration for ignoring variables
func Analyze(codeUsages []EnvUsage, envVars map[string]string, envVarsFromFiles map[string]string, envKeySources map[string]string, cfg *config.Config) ScanResult {
	result := ScanResult{
		CodeKeys:            codeUsages,
		EnvKeys:             envVarsFromFiles, // Store .env file vars for display purposes
		EnvKeySources:       envKeySources,    // Store source file for each variable
		Missing:             make(map[string][]EnvUsage),
		PartialMatches:      make(map[string][]EnvUsage),
		Unused:              []string{},
		IgnoredMissing:      0,
		IgnoredFromFolders:  0,
	}

	// Build a map of keys used in code, separating full and partial matches
	codeKeys := make(map[string][]EnvUsage)
	partialKeys := make(map[string][]EnvUsage)
	for _, usage := range codeUsages {
		if usage.IsPartial {
			// For partial matches with a full expression, use the full expression as the key
			// This ensures we group by the actual expression and display it correctly
			key := usage.Key
			if usage.FullExpr != "" {
				key = usage.FullExpr
			}
			partialKeys[key] = append(partialKeys[key], usage)
		} else {
			codeKeys[usage.Key] = append(codeKeys[usage.Key], usage)
		}
	}

	// Track unique variables from ignored folders that would have been missing
	ignoredFolderVars := make(map[string]bool)

	// Find missing keys (in code but not in envVars - checks both .env and exported env)
	// Filter out ignored variables and variables from ignored folders
	for key, usages := range codeKeys {
		if _, exists := envVars[key]; !exists {
			// Check if all usages are from ignored folders
			allInIgnoredFolders := true
			hasIgnoredFolderUsage := false
			for _, usage := range usages {
				if usage.InIgnoredPath {
					hasIgnoredFolderUsage = true
				} else {
					allInIgnoredFolders = false
				}
			}
			
			// If all usages are from ignored folders, count it but don't report as missing
			if allInIgnoredFolders && hasIgnoredFolderUsage {
				ignoredFolderVars[key] = true
				continue
			}
			
			// Check if this variable should be ignored via config
			if cfg != nil && cfg.ShouldIgnoreMissing(key) {
				result.IgnoredMissing++
			} else {
				// Only include usages that are NOT from ignored folders
				var nonIgnoredUsages []EnvUsage
				for _, usage := range usages {
					if !usage.InIgnoredPath {
						nonIgnoredUsages = append(nonIgnoredUsages, usage)
					}
				}
				if len(nonIgnoredUsages) > 0 {
					result.Missing[key] = nonIgnoredUsages
				}
			}
		}
	}
	
	// Count unique variables from ignored folders
	result.IgnoredFromFolders = len(ignoredFolderVars)

	// Handle partial matches - check if any env vars contain the partial string
	for key, usages := range partialKeys {
		// Check if this is a variable reference pattern (e.g., process.env[a])
		// These should always be reported as partial matches since we can't determine
		// the actual env var name at static analysis time
		isVarRef := false
		for _, usage := range usages {
			if usage.IsVarRef {
				isVarRef = true
				break
			}
		}
		
		if isVarRef {
			// Always report variable reference patterns as partial matches
			result.PartialMatches[key] = usages
			continue
		}
		
		// For string-based partial matches, check if any env vars contain the partial string
		hasMatch := false
		for envKey := range envVars {
			// Check if any env var contains the partial string
			// This works for prefix patterns (e.g., "MY_" from "MY_" + var)
			// suffix patterns (e.g., "_VAR" from var + "_VAR")
			// and middle patterns (e.g., "fff" from "asdf" + var + "fff")
			if strings.Contains(envKey, key) {
				hasMatch = true
				break
			}
		}
		
		// If no match found, add to partial matches
		if !hasMatch {
			result.PartialMatches[key] = usages
		}
	}

	// Find unused keys (in .env files but not in code)
	// Only check envVarsFromFiles, not exported environment variables
	for key := range envVarsFromFiles {
		if _, exists := codeKeys[key]; !exists {
			result.Unused = append(result.Unused, key)
		}
	}

	return result
}

