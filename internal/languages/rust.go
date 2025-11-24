package languages

// RustQuery is the Tree-Sitter query for finding env::var("KEY") and std::env::var("KEY") patterns
// Also supports dynamic patterns like env::var("prefix_" + var) and env::var(var)
// Note: We don't use predicates here, filtering is done in ExtractEnvVarsFromRust
const RustQuery = `
[
  (call_expression
    function: (scoped_identifier
      path: (identifier) @path
      name: (identifier) @fn
    )
    arguments: (arguments (string_literal) @key)
  )
  (call_expression
    function: (scoped_identifier
      path: (scoped_identifier
        path: (identifier) @path1
        name: (identifier) @path2
      )
      name: (identifier) @fn
    )
    arguments: (arguments (string_literal) @key)
  )
  (call_expression
    function: (scoped_identifier
      path: (identifier) @path
      name: (identifier) @fn
    )
    arguments: (arguments (binary_expression) @full_expr)
  )
  (call_expression
    function: (scoped_identifier
      path: (scoped_identifier
        path: (identifier) @path1
        name: (identifier) @path2
      )
      name: (identifier) @fn
    )
    arguments: (arguments (binary_expression) @full_expr)
  )
  (call_expression
    function: (scoped_identifier
      path: (identifier) @path
      name: (identifier) @fn
    )
    arguments: (arguments (identifier) @var)
  )
  (call_expression
    function: (scoped_identifier
      path: (scoped_identifier
        path: (identifier) @path1
        name: (identifier) @path2
      )
      name: (identifier) @fn
    )
    arguments: (arguments (identifier) @var)
  )
]
`

// ExtractEnvVarsFromRust extracts environment variable keys from Rust AST matches
// Returns []string for backward compatibility
func ExtractEnvVarsFromRust(matches []map[string]string) []string {
	results := ExtractEnvVarsFromRustWithPartial(matches)
	var keys []string
	for _, result := range results {
		if !result.IsPartial {
			keys = append(keys, result.Key)
		}
	}
	return keys
}

// ExtractEnvVarsFromRustWithPartial extracts environment variable keys from Rust AST matches
// Returns matches with partial match information
func ExtractEnvVarsFromRustWithPartial(matches []map[string]string) []EnvVarMatch {
	var results []EnvVarMatch
	seen := make(map[string]bool)

	for _, match := range matches {
		fn, fnOk := match["fn"]
		path, pathOk := match["path"]
		path1, path1Ok := match["path1"]
		path2, path2Ok := match["path2"]

		if !fnOk {
			continue
		}

		// Validate function name
		if fn != "var" && fn != "var_os" {
			continue
		}

		// Validate path: either "env" or "std::env"
		isValidPath := false
		if path1Ok && path2Ok {
			isValidPath = path1 == "std" && path2 == "env"
		} else if pathOk {
			isValidPath = path == "env"
		}

		if !isValidPath {
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

		// Case 2: Binary expression (e.g., "prefix_" + var, var + "_suffix")
		fullExpr, fullExprOk := match["full_expr"]
		if fullExprOk && fullExpr != "" {
			if !seen[fullExpr] {
				results = append(results, EnvVarMatch{
					Key:       fullExpr,
					IsPartial: true,
					FullExpr:  fullExpr,
				})
				seen[fullExpr] = true
			}
			continue
		}

		// Case 3: Variable identifier (e.g., env::var(var))
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

