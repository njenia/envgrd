package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jenian/envgrd/internal/analyzer"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// JSONOutput represents the JSON output format
type JSONOutput struct {
	Missing            []MissingVar `json:"missing"`
	PartialMatches     []MissingVar `json:"partial_matches"`
	Unused             []string     `json:"unused"`
	IgnoredMissing     int          `json:"ignored_missing"`
	IgnoredFromFolders int          `json:"ignored_from_folders"`
}

// MissingVar represents a missing environment variable with its locations
type MissingVar struct {
	Key       string   `json:"key"`
	Locations []string `json:"locations"`
}

// Format formats the scan results according to the specified format
func Format(result analyzer.ScanResult, jsonOutput bool, silent bool, skipUnused bool, dynamic bool) error {
	if silent {
		// In silent mode, only return exit code (handled by caller)
		return nil
	}

	if jsonOutput {
		return formatJSON(result, skipUnused, dynamic)
	}

	return formatHumanReadable(result, skipUnused, dynamic)
}

// formatJSON outputs results in JSON format
func formatJSON(result analyzer.ScanResult, skipUnused bool, dynamic bool) error {
	output := JSONOutput{
		Missing:            []MissingVar{},
		PartialMatches:     []MissingVar{},
		Unused:             []string{},
		IgnoredMissing:     result.IgnoredMissing,
		IgnoredFromFolders: result.IgnoredFromFolders,
	}

	// Convert missing vars
	for key, usages := range result.Missing {
		locations := make([]string, 0, len(usages))
		for _, usage := range usages {
			loc := fmt.Sprintf("%s:%d", usage.File, usage.Line)
			if usage.CodeSnippet != "" {
				loc += fmt.Sprintf(" (%s)", usage.CodeSnippet)
			}
			locations = append(locations, loc)
		}
		sort.Strings(locations)
		output.Missing = append(output.Missing, MissingVar{
			Key:       key,
			Locations: locations,
		})
	}

	// Sort missing vars by key
	sort.Slice(output.Missing, func(i, j int) bool {
		return output.Missing[i].Key < output.Missing[j].Key
	})

	// Convert partial matches
	for key, usages := range result.PartialMatches {
		locations := make([]string, 0, len(usages))
		for _, usage := range usages {
			loc := fmt.Sprintf("%s:%d", usage.File, usage.Line)
			if usage.CodeSnippet != "" {
				loc += fmt.Sprintf(" (%s)", usage.CodeSnippet)
			}
			locations = append(locations, loc)
		}
		sort.Strings(locations)
		output.PartialMatches = append(output.PartialMatches, MissingVar{
			Key:       key,
			Locations: locations,
		})
	}

	// Sort partial matches by key
	sort.Slice(output.PartialMatches, func(i, j int) bool {
		return output.PartialMatches[i].Key < output.PartialMatches[j].Key
	})

	// Only include partial matches if dynamic mode is enabled
	if !dynamic {
		output.PartialMatches = []MissingVar{}
	}

	// Add unused vars if not skipped
	if !skipUnused {
		output.Unused = make([]string, len(result.Unused))
		copy(output.Unused, result.Unused)
		sort.Strings(output.Unused)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatHumanReadable outputs results in human-readable format
func formatHumanReadable(result analyzer.ScanResult, skipUnused bool, dynamic bool) error {
	hasIssues := false

	// Missing variables
	if len(result.Missing) > 0 {
		hasIssues = true
		fmt.Printf("%s%sMissing environment variables:%s\n\n", colorBold, colorRed, colorReset)
		keys := make([]string, 0, len(result.Missing))
		for key := range result.Missing {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			usages := result.Missing[key]
			fmt.Printf("  %s%s%s\n", colorRed, key, colorReset)
			for _, usage := range usages {
				filePath := usage.File
				if filePath == "" {
					filePath = "<unknown>"
				}
				fmt.Printf("    %sused in:%s %s%s%s:%s%d%s", colorGray, colorReset, colorCyan, filePath, colorReset, colorYellow, usage.Line, colorReset)
				if usage.CodeSnippet != "" {
					// Truncate long snippets
					snippet := usage.CodeSnippet
					if len(snippet) > 80 {
						snippet = snippet[:77] + "..."
					}
					fmt.Printf(" %s%s%s", colorGray, snippet, colorReset)
				}
				fmt.Println()
			}
			fmt.Println()
		}
	}

	// Partial matches (dynamic patterns) - only show if dynamic mode is enabled
	if dynamic && len(result.PartialMatches) > 0 {
		hasIssues = true
		fmt.Printf("%s%sDynamic patterns (runtime-evaluated expressions):%s\n", colorBold, colorYellow, colorReset)
		keys := make([]string, 0, len(result.PartialMatches))
		for key := range result.PartialMatches {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			usages := result.PartialMatches[key]
			// Display the key directly (which is the full expression for dynamic patterns)
			fmt.Printf("  %s%s%s\n", colorYellow, key, colorReset)
			for _, usage := range usages {
				filePath := usage.File
				if filePath == "" {
					filePath = "<unknown>"
				}
				fmt.Printf("    %sused in:%s %s%s%s:%s%d%s", colorGray, colorReset, colorCyan, filePath, colorReset, colorYellow, usage.Line, colorReset)
				if usage.CodeSnippet != "" {
					// Truncate long snippets
					snippet := usage.CodeSnippet
					if len(snippet) > 80 {
						snippet = snippet[:77] + "..."
					}
					fmt.Printf(" %s%s%s", colorGray, snippet, colorReset)
				}
				fmt.Println()
			}
			fmt.Println()
		}
	}

	// Unused variables
	if !skipUnused && len(result.Unused) > 0 {
		hasIssues = true
		fmt.Printf("%s%sUnused variables:%s\n\n", colorBold, colorYellow, colorReset)
		sort.Strings(result.Unused)
		for _, key := range result.Unused {
			value := result.EnvKeys[key]
			// Redact the value
			redactedValue := redactValue(value)
			fmt.Printf("  %s%s%s=%s%s%s %s(in .env)%s\n", colorYellow, key, colorReset, colorGray, redactedValue, colorReset, colorGray, colorReset)
		}
		fmt.Println()
	}

	// Show ignored missing variables count
	if result.IgnoredMissing > 0 {
		fmt.Printf("%s%sNote:%s %d missing variable(s) were ignored (configured in .envgrd.config)\n", colorGray, colorBold, colorReset, result.IgnoredMissing)
	}

	// Show ignored variables from ignored folders
	if result.IgnoredFromFolders > 0 {
		fmt.Printf("%s%sNote:%s %d variable(s) found in ignored folders were excluded from the scan (configured in .envgrd.config)\n", colorGray, colorBold, colorReset, result.IgnoredFromFolders)
	}

	if result.IgnoredMissing > 0 || result.IgnoredFromFolders > 0 {
		fmt.Println()
	}

	// No issues found
	if !hasIssues {
		ignoredCount := result.IgnoredMissing + result.IgnoredFromFolders
		if ignoredCount > 0 {
			var parts []string
			if result.IgnoredMissing > 0 {
				parts = append(parts, fmt.Sprintf("%d ignored via config", result.IgnoredMissing))
			}
			if result.IgnoredFromFolders > 0 {
				parts = append(parts, fmt.Sprintf("%d from ignored folders", result.IgnoredFromFolders))
			}
			fmt.Printf("%s%s✓ No issues found (excluding %s).%s\n", colorGreen, colorBold, strings.Join(parts, ", "), colorReset)
		} else {
			fmt.Printf("%s%s✓ No issues found. All environment variables are properly configured.%s\n", colorGreen, colorBold, colorReset)
		}
	}

	return nil
}

// redactValue redacts sensitive values while showing the type
func redactValue(value string) string {
	if value == "" {
		return `""`
	}
	// If it looks like a secret (long, random-looking), redact it
	if len(value) > 20 {
		return "[REDACTED]"
	}
	// If it contains special characters that suggest it's a secret
	if strings.ContainsAny(value, "=+/") && len(value) > 10 {
		return "[REDACTED]"
	}
	// For short values, show first and last char
	if len(value) > 4 {
		return string(value[0]) + "..." + string(value[len(value)-1])
	}
	// For very short values, just show asterisks
	return "***"
}

// HasIssues returns true if there are any issues in the scan result
// Note: Ignored missing variables don't count as issues
// dynamic: whether to include partial matches in the issue count
func HasIssues(result analyzer.ScanResult, skipUnused bool, dynamic bool) bool {
	if len(result.Missing) > 0 {
		return true
	}
	if dynamic && len(result.PartialMatches) > 0 {
		return true
	}
	if !skipUnused && len(result.Unused) > 0 {
		return true
	}
	return false
}

// FormatError formats an error message
func FormatError(err error) string {
	return fmt.Sprintf("Error: %s\n", err)
}
