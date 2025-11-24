package languages

// JavaScriptQuery is the Tree-Sitter query for finding process.env.KEY patterns
// Supports both dot notation (process.env.KEY) and bracket notation (process.env["KEY"])
// Also supports partial matches for dynamic patterns (process.env["prefix_" + var])
// Note: We don't use predicates here, filtering is done in ExtractEnvVarsFromJS
const JavaScriptQuery = `
[
  (member_expression
    object: (member_expression
      object: (identifier) @obj
      property: (property_identifier) @prop
    )
    property: (property_identifier) @key
  )
  (subscript_expression
    object: (member_expression
      object: (identifier) @obj
      property: (property_identifier) @prop
    )
    index: (string) @key
  )
  (subscript_expression
    object: (member_expression
      object: (identifier) @obj
      property: (property_identifier) @prop
    )
    index: (binary_expression) @full_expr
  )
  (subscript_expression
    object: (member_expression
      object: (identifier) @obj
      property: (property_identifier) @prop
    )
    index: (identifier) @var
  )
]
`

// ExtractEnvVarsFromJS extracts environment variable keys from JavaScript/TypeScript AST matches
// Returns matches with partial match information
func ExtractEnvVarsFromJS(matches []map[string]string) []EnvVarMatch {
	var results []EnvVarMatch
	seen := make(map[string]bool)

	for _, match := range matches {
		// Validate that this is actually process.env
		obj, objOk := match["obj"]
		prop, propOk := match["prop"]

		if !objOk || !propOk || obj != "process" || prop != "env" {
			continue
		}

		// Case 1: Static key (dot notation or bracket notation with string literal)
		key, keyOk := match["key"]
		if keyOk && key != "" {
			// Remove quotes if present
			key = trimQuotes(key)
			if key != "" && !seen[key] {
				results = append(results, EnvVarMatch{Key: key, IsPartial: false})
				seen[key] = true
			}
			continue
		}

		// Case 2 & 3: Partial match - binary expression (e.g., "prefix_" + var, var + "_suffix", "asdf" + var + "fff")
		fullExpr, fullExprOk := match["full_expr"]
		if fullExprOk && fullExpr != "" {
			// Extract string parts from the expression for matching
			// The full expression is stored for display, but we extract string parts for matching
			// Try to find the first or last string literal in the expression
			firstStr := extractFirstString(fullExpr)
			lastStr := extractLastString(fullExpr)

			var key string
			var displayKey string

			if firstStr != "" && lastStr != "" && firstStr == lastStr {
				// Single string part (e.g., "prefix_" + var or var + "_suffix")
				key = firstStr + "*"
				displayKey = firstStr
			} else if firstStr != "" {
				// String at the start (e.g., "prefix_" + var)
				key = firstStr + "*"
				displayKey = firstStr
			} else if lastStr != "" {
				// String at the end (e.g., var + "_suffix")
				key = "*" + lastStr
				displayKey = lastStr
			} else {
				// No string parts found - use full expression
				key = fullExpr
				displayKey = fullExpr
			}

			if key != "" && !seen[key] {
				results = append(results, EnvVarMatch{
					Key:       displayKey,
					IsPartial: true,
					FullExpr:  fullExpr,
				})
				seen[key] = true
			}
			continue
		}

		// Case 4: Partial match - variable identifier (e.g., process.env[a])
		varName, varOk := match["var"]
		if varOk && varName != "" {
			// This is a dynamic pattern - we can't determine the actual env var name
			// Report it as a partial match with the variable name
			// Use a special prefix to distinguish from string-based partial matches
			key := "[var:" + varName + "]"
			if !seen[key] {
				results = append(results, EnvVarMatch{Key: varName, IsPartial: true, IsVarRef: true})
				seen[key] = true
			}
		}
	}

	return results
}

// extractFirstString extracts the first string literal from an expression
func extractFirstString(expr string) string {
	// Look for the first quoted string in the expression
	// Simple regex-like approach: find "..." or '...' or `...`
	start := -1
	var quote byte
	for i := 0; i < len(expr); i++ {
		if expr[i] == '"' || expr[i] == '\'' || expr[i] == '`' {
			if start == -1 {
				start = i
				quote = expr[i]
			} else if expr[i] == quote {
				// Found matching quote
				return expr[start+1 : i]
			}
		}
	}
	return ""
}

// extractLastString extracts the last string literal from an expression
func extractLastString(expr string) string {
	// Look for the last quoted string in the expression
	end := -1
	var quote byte
	for i := len(expr) - 1; i >= 0; i-- {
		if expr[i] == '"' || expr[i] == '\'' || expr[i] == '`' {
			if end == -1 {
				end = i
				quote = expr[i]
			} else if expr[i] == quote {
				// Found matching quote
				return expr[i+1 : end]
			}
		}
	}
	return ""
}
