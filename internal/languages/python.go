package languages

// PythonQuery is the Tree-Sitter query for finding os.environ["KEY"] and os.getenv("KEY") patterns
// Also supports dynamic patterns like os.environ["prefix_" + var] and os.getenv(var)
// Note: We don't use predicates here, filtering is done in ExtractEnvVarsFromPython
const PythonQuery = `
[
  (subscript
    value: (attribute
      object: (identifier) @obj
      attribute: (identifier) @attr
    )
    subscript: (string) @key
  )
  (call
    function: (attribute
      object: (identifier) @obj2
      attribute: (identifier) @fn
    )
    arguments: (argument_list (string) @key)
  )
  (subscript
    value: (attribute
      object: (identifier) @obj
      attribute: (identifier) @attr
    )
    subscript: (binary_operator) @full_expr
  )
  (subscript
    value: (attribute
      object: (identifier) @obj
      attribute: (identifier) @attr
    )
    subscript: (identifier) @var
  )
  (call
    function: (attribute
      object: (identifier) @obj2
      attribute: (identifier) @fn
    )
    arguments: (argument_list (binary_operator) @full_expr)
  )
  (call
    function: (attribute
      object: (identifier) @obj2
      attribute: (identifier) @fn
    )
    arguments: (argument_list (identifier) @var)
  )
]
`

// ExtractEnvVarsFromPython extracts environment variable keys from Python AST matches
// Returns []string for backward compatibility
func ExtractEnvVarsFromPython(matches []map[string]string) []string {
	results := ExtractEnvVarsFromPythonWithPartial(matches)
	var keys []string
	for _, result := range results {
		if !result.IsPartial {
			keys = append(keys, result.Key)
		}
	}
	return keys
}

// ExtractEnvVarsFromPythonWithPartial extracts environment variable keys from Python AST matches
// Returns matches with partial match information
func ExtractEnvVarsFromPythonWithPartial(matches []map[string]string) []EnvVarMatch {
	var results []EnvVarMatch
	seen := make(map[string]bool)

	for _, match := range matches {
		key, keyOk := match["key"]
		obj, objOk := match["obj"]
		attr, attrOk := match["attr"]
		fn, fnOk := match["fn"]
		obj2, obj2Ok := match["obj2"]

		// Check for os.environ["KEY"] pattern
		if keyOk && objOk && attrOk && key != "" {
			if obj == "os" && attr == "environ" {
				key = trimQuotes(key)
				if key != "" && !seen[key] {
					results = append(results, EnvVarMatch{Key: key, IsPartial: false})
					seen[key] = true
				}
				continue
			}
		}

		// Check for os.getenv("KEY") pattern
		if keyOk && obj2Ok && fnOk && key != "" {
			if obj2 == "os" && fn == "getenv" {
				key = trimQuotes(key)
				if key != "" && !seen[key] {
					results = append(results, EnvVarMatch{Key: key, IsPartial: false})
					seen[key] = true
				}
				continue
			}
		}

		// Case 2: Binary expression for os.environ["prefix_" + var]
		fullExpr, fullExprOk := match["full_expr"]
		if fullExprOk && fullExpr != "" {
			// Validate it's os.environ or os.getenv
			isValid := false
			if objOk && attrOk && obj == "os" && attr == "environ" {
				isValid = true
			} else if obj2Ok && fnOk && obj2 == "os" && fn == "getenv" {
				isValid = true
			}

			if isValid && !seen[fullExpr] {
				results = append(results, EnvVarMatch{
					Key:       fullExpr,
					IsPartial: true,
					FullExpr:  fullExpr,
				})
				seen[fullExpr] = true
			}
			continue
		}

		// Case 3: Variable identifier (e.g., os.environ[var] or os.getenv(var))
		varName, varOk := match["var"]
		if varOk && varName != "" {
			// Validate it's os.environ or os.getenv
			isValid := false
			if objOk && attrOk && obj == "os" && attr == "environ" {
				isValid = true
			} else if obj2Ok && fnOk && obj2 == "os" && fn == "getenv" {
				isValid = true
			}

			if isValid && !seen[varName] {
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

