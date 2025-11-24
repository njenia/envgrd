package languages

// GoQuery is the Tree-Sitter query for finding os.Getenv("KEY") patterns
// Also supports dynamic patterns like os.Getenv("prefix_" + var) and os.Getenv(var)
// Note: We don't use predicates here, filtering is done in ExtractEnvVarsFromGo
const GoQuery = `
[
  (call_expression
    function: (selector_expression
      operand: (identifier) @obj
      field: (field_identifier) @fn
    )
    arguments: (argument_list (interpreted_string_literal) @key)
  )
  (call_expression
    function: (selector_expression
      operand: (identifier) @obj
      field: (field_identifier) @fn
    )
    arguments: (argument_list (binary_expression) @full_expr)
  )
  (call_expression
    function: (selector_expression
      operand: (identifier) @obj
      field: (field_identifier) @fn
    )
    arguments: (argument_list (identifier) @var)
  )
]
`

// ExtractEnvVarsFromGo extracts environment variable keys from Go AST matches
// Returns []string for backward compatibility
func ExtractEnvVarsFromGo(matches []map[string]string) []string {
	results := ExtractEnvVarsFromGoWithPartial(matches)
	var keys []string
	for _, result := range results {
		if !result.IsPartial {
			keys = append(keys, result.Key)
		}
	}
	return keys
}

// ExtractEnvVarsFromGoWithPartial extracts environment variable keys from Go AST matches
// Returns matches with partial match information
func ExtractEnvVarsFromGoWithPartial(matches []map[string]string) []EnvVarMatch {
	var results []EnvVarMatch
	seen := make(map[string]bool)

	for _, match := range matches {
		// Validate that this is actually os.Getenv
		obj, objOk := match["obj"]
		fn, fnOk := match["fn"]

		if !objOk || !fnOk || obj != "os" || fn != "Getenv" {
			continue
		}

		// Case 1: Static key (string literal)
		key, keyOk := match["key"]
		if keyOk && key != "" {
			key = trimQuotes(key)
			if key != "" && !seen[key] {
				results = append(results, EnvVarMatch{Key: key, IsPartial: false})
				seen[key] = true
			}
			continue
		}

		// Case 2: Binary expression (e.g., "prefix_" + var, var + "_suffix", "asdf" + var + "fff")
		fullExpr, fullExprOk := match["full_expr"]
		if fullExprOk && fullExpr != "" {
			if !seen[fullExpr] {
				// Use FullExpr as the key for grouping and display
				results = append(results, EnvVarMatch{
					Key:       fullExpr, // Use full expression as key for display
					IsPartial: true,
					FullExpr:  fullExpr,
				})
				seen[fullExpr] = true
			}
			continue
		}

		// Case 3: Variable identifier (e.g., os.Getenv(var))
		varName, varOk := match["var"]
		if varOk && varName != "" {
			if !seen[varName] {
				results = append(results, EnvVarMatch{
					Key:       varName,
					IsPartial: true,
					IsVarRef:  true,
				})
				seen[varName] = true
			}
		}
	}

	return results
}

// trimQuotes removes surrounding quotes from a string
func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '`' && s[len(s)-1] == '`') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
