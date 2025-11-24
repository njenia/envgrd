package languages

// JavaQuery is the Tree-Sitter query for finding System.getenv("KEY") and System.getenv().get("KEY") patterns
// Also supports dynamic patterns like System.getenv("prefix_" + var) and System.getenv(var)
// Note: We don't use predicates here, filtering is done in ExtractEnvVarsFromJava
const JavaQuery = `
[
  (method_invocation
    object: (identifier) @obj
    name: (identifier) @method
    arguments: (argument_list (string_literal) @key)
  )
  (method_invocation
    object: (method_invocation
      object: (identifier) @obj
      name: (identifier) @method1
    )
    name: (identifier) @method2
    arguments: (argument_list (string_literal) @key)
  )
  (method_invocation
    object: (identifier) @obj
    name: (identifier) @method
    arguments: (argument_list (binary_expression) @full_expr)
  )
  (method_invocation
    object: (method_invocation
      object: (identifier) @obj
      name: (identifier) @method1
    )
    name: (identifier) @method2
    arguments: (argument_list (binary_expression) @full_expr)
  )
  (method_invocation
    object: (identifier) @obj
    name: (identifier) @method
    arguments: (argument_list (identifier) @var)
  )
  (method_invocation
    object: (method_invocation
      object: (identifier) @obj
      name: (identifier) @method1
    )
    name: (identifier) @method2
    arguments: (argument_list (identifier) @var)
  )
]
`

// ExtractEnvVarsFromJava extracts environment variable keys from Java AST matches
// Returns []string for backward compatibility
func ExtractEnvVarsFromJava(matches []map[string]string) []string {
	results := ExtractEnvVarsFromJavaWithPartial(matches)
	var keys []string
	for _, result := range results {
		if !result.IsPartial {
			keys = append(keys, result.Key)
		}
	}
	return keys
}

// ExtractEnvVarsFromJavaWithPartial extracts environment variable keys from Java AST matches
// Returns matches with partial match information
func ExtractEnvVarsFromJavaWithPartial(matches []map[string]string) []EnvVarMatch {
	var results []EnvVarMatch
	seen := make(map[string]bool)

	for _, match := range matches {
		obj, objOk := match["obj"]
		method, methodOk := match["method"]
		method1, method1Ok := match["method1"]
		method2, method2Ok := match["method2"]

		if !objOk || obj != "System" {
			continue
		}

		// Validate method calls
		isValidCall := false
		if methodOk && method == "getenv" {
			isValidCall = true
		} else if method1Ok && method2Ok && method1 == "getenv" && method2 == "get" {
			isValidCall = true
		}

		if !isValidCall {
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

		// Case 3: Variable identifier (e.g., System.getenv(var))
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

