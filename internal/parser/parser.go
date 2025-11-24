package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jenian/envgrd/internal/analyzer"
	"github.com/jenian/envgrd/internal/languages"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Parser handles Tree-Sitter parsing of source files
type Parser struct {
	languages map[string]*sitter.Language
	mu        sync.RWMutex
	debug     bool
}


// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{
		languages: make(map[string]*sitter.Language),
		debug:     false,
	}
}

// SetDebug enables or disables debug logging
func (p *Parser) SetDebug(debug bool) {
	p.debug = debug
}

// getLanguage returns a language grammar for the given language, loading it if needed
func (p *Parser) getLanguage(lang string) (*sitter.Language, error) {
	p.mu.RLock()
	if language, ok := p.languages[lang]; ok {
		p.mu.RUnlock()
		return language, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if language, ok := p.languages[lang]; ok {
		return language, nil
	}

	// Load language grammar
	language, err := loadLanguage(lang)
	if err != nil {
		return nil, fmt.Errorf("failed to load language %s: %w", lang, err)
	}

	p.languages[lang] = language
	return language, nil
}


// ParseFile parses a single file and extracts environment variable usages
// scanRoot is the root directory being scanned, used for calculating relative paths
func (p *Parser) ParseFile(filePath string, lang string, scanRoot string) ([]analyzer.EnvUsage, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Get language grammar
	language, err := p.getLanguage(lang)
	if err != nil {
		if p.debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Failed to load language %s for %s: %v\n", lang, filePath, err)
		}
		return nil, err
	}
	if language == nil {
		if p.debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Language is nil for %s (language: %s)\n", filePath, lang)
		}
		return []analyzer.EnvUsage{}, nil
	}

	// Parse the file using the official tree-sitter API
	// Create a new parser for each file to avoid CGO concurrency issues
	// Tree-sitter parsers are not thread-safe when used concurrently
	tsParser := sitter.NewParser()
	defer tsParser.Close()
	tsParser.SetLanguage(language)
	
	var rootNode *sitter.Node
	tree := tsParser.Parse(content, nil)
	if tree != nil {
		rootNode = tree.RootNode()
		defer tree.Close()
	} else {
		if p.debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Parse returned nil tree for %s (language: %s)\n", filePath, lang)
		}
	}
	
	// If still nil, return empty results (parsing failed)
	if rootNode == nil {
		if p.debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] RootNode is nil for %s (language: %s)\n", filePath, lang)
		}
		return []analyzer.EnvUsage{}, nil
	}

	// Get language-specific query and extractor
	langInfo := languages.GetLanguageInfo(lang)
	if langInfo == nil {
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}

	// Create query - trim whitespace to avoid parsing issues
	queryStr := strings.TrimSpace(langInfo.Query)
	if queryStr == "" {
		return nil, fmt.Errorf("empty query for language: %s", lang)
	}
	
	query, queryErr := sitter.NewQuery(language, queryStr)
	if queryErr != nil {
		// Query creation failed - this might be due to grammar compatibility
		// Log the error but return empty results to allow scan to continue
		if p.debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Query creation failed for %s: %v\n", filePath, queryErr)
			fmt.Fprintf(os.Stderr, "[DEBUG] Query was: %s\n", queryStr)
			// Try to get some info about the parsed tree
			if rootNode != nil {
				fmt.Fprintf(os.Stderr, "[DEBUG] Root node type: %s, children: %d\n", rootNode.GrammarName(), rootNode.ChildCount())
			}
		}
		return []analyzer.EnvUsage{}, nil
	}
	defer query.Close()

	// Execute query using QueryCursor
	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	matches := cursor.Matches(query, rootNode, content)

	// Collect matches with node information
	type matchInfo struct {
		key         string
		node        *sitter.Node
		codeSnippet string
		isPartial   bool
		isVarRef    bool
		fullExpr    string
	}
	var matchInfos []matchInfo

	// Get capture names for lookup
	captureNames := query.CaptureNames()

	for {
		match := matches.Next()
		if match == nil {
			break
		}

		matchMap := make(map[string]string)
		var keyNode *sitter.Node
		var objNode *sitter.Node
		var propNode *sitter.Node
		var fullMatchNode *sitter.Node
		var leftStrNode *sitter.Node
		var rightStrNode *sitter.Node
		var varNode *sitter.Node
		var fullExprNode *sitter.Node

		for _, capture := range match.Captures {
			// Get capture name from index
			captureIndex := capture.Index
			if int(captureIndex) < len(captureNames) {
				captureName := captureNames[captureIndex]
				captureNode := &capture.Node // Convert to pointer
				captureText := string(content[captureNode.StartByte():captureNode.EndByte()])
				matchMap[captureName] = captureText

				switch captureName {
				case "key":
					keyNode = captureNode
				case "obj":
					objNode = captureNode
				case "prop":
					propNode = captureNode
				case "left_str":
					leftStrNode = captureNode
				case "right_str":
					rightStrNode = captureNode
				case "var":
					varNode = captureNode
				case "full_expr":
					fullExprNode = captureNode
				}

				// Get the full member_expression/subscript_expression node for context
				if captureName == "key" || captureName == "left_str" || captureName == "right_str" || captureName == "var" || captureName == "full_expr" {
					// Use the match node itself for context
					if fullMatchNode == nil {
						fullMatchNode = captureNode
					}
				}
			}
		}

		// Extract keys from this match
		// For JavaScript/TypeScript, use the special extractor that returns partial match info
		var matches []languages.EnvVarMatch
		if langInfo.ExtractorWithPartial != nil {
			matches = langInfo.ExtractorWithPartial([]map[string]string{matchMap})
		} else if langInfo.Extractor != nil {
			// For other languages, convert string results to EnvVarMatch
			keys := langInfo.Extractor([]map[string]string{matchMap})
			for _, key := range keys {
				matches = append(matches, languages.EnvVarMatch{Key: key, IsPartial: false})
			}
		}
		
		for _, match := range matches {
			key := match.Key
			isPartial := match.IsPartial
			
			// Determine which node to use for line number and context
			var nodeForContext *sitter.Node
			if isPartial {
				// For partial matches, prefer the full expression node, then string node, then var node
				if fullExprNode != nil {
					nodeForContext = fullExprNode
				} else if leftStrNode != nil {
					nodeForContext = leftStrNode
				} else if rightStrNode != nil {
					nodeForContext = rightStrNode
				} else if varNode != nil {
					nodeForContext = varNode
				} else {
					nodeForContext = keyNode
				}
			} else {
				nodeForContext = keyNode
			}
			
			// For variable references, if we don't have a specific node, use the full match node
			if nodeForContext == nil && match.IsVarRef && fullMatchNode != nil {
				nodeForContext = fullMatchNode
			}
			
			if nodeForContext != nil {
				// Get code context around the match
				startByte := nodeForContext.StartByte()
				endByte := nodeForContext.EndByte()
				if fullMatchNode != nil {
					startByte = fullMatchNode.StartByte()
					endByte = fullMatchNode.EndByte()
				}

				// Get surrounding context (100 chars before and after)
				contextStart := int(startByte) - 100
				if contextStart < 0 {
					contextStart = 0
				}
				contextEnd := int(endByte) + 100
				if contextEnd > len(content) {
					contextEnd = len(content)
				}

				// Get code snippet from the line
				startPos := nodeForContext.StartPosition()
				lineNum := int(startPos.Row)
				lineStart := 0
				for i := 0; i < len(content) && lineNum > 0; i++ {
					if content[i] == '\n' {
						lineNum--
						lineStart = i + 1
					}
				}
				lineEnd := lineStart
				for lineEnd < len(content) && content[lineEnd] != '\n' {
					lineEnd++
				}
				codeSnippet := string(content[lineStart:lineEnd])
				// Trim whitespace
				codeSnippet = strings.TrimSpace(codeSnippet)

				// Log the match for debugging (only if debug is enabled)
				if p.debug {
					line := int(startPos.Row) + 1
					fullText := string(content[startByte:endByte])
					context := string(content[contextStart:contextEnd])
					fmt.Fprintf(os.Stderr, "[DEBUG] Match in %s:%d\n", filePath, line)
					fmt.Fprintf(os.Stderr, "  Full match: %q\n", fullText)
					fmt.Fprintf(os.Stderr, "  Extracted key: %q\n", key)
					if objNode != nil {
						fmt.Fprintf(os.Stderr, "  Object: %q\n", string(content[objNode.StartByte():objNode.EndByte()]))
					}
					if propNode != nil {
						fmt.Fprintf(os.Stderr, "  Property: %q\n", string(content[propNode.StartByte():propNode.EndByte()]))
					}
					fmt.Fprintf(os.Stderr, "  Context: %q\n", context)
					fmt.Fprintf(os.Stderr, "  ---\n")
				}

				matchInfos = append(matchInfos, matchInfo{
					key:         key,
					node:        nodeForContext,
					codeSnippet: codeSnippet,
					isPartial:   isPartial,
					isVarRef:    match.IsVarRef,
					fullExpr:    match.FullExpr,
				})
			}
		}
	}

	// Convert to EnvUsage with line numbers
	var usages []analyzer.EnvUsage
	seen := make(map[string]bool)

	// Use relative path from scanRoot if possible, otherwise use the provided path
	relPath := filePath
	if scanRoot != "" {
		// Make both paths absolute for comparison
		absScanRoot, err1 := filepath.Abs(scanRoot)
		absFilePath, err2 := filepath.Abs(filePath)
		if err1 == nil && err2 == nil {
			if rel, err := filepath.Rel(absScanRoot, absFilePath); err == nil && rel != "" {
				relPath = rel
			}
		}
	}
	
	// Fallback: if relPath is still empty or invalid, use filePath
	if relPath == "" {
		relPath = filePath
	}

	for _, matchInfo := range matchInfos {
		// Get line number from node (1-indexed)
		startPos := matchInfo.node.StartPosition()
		line := int(startPos.Row) + 1

		usageKey := fmt.Sprintf("%s:%s:%d", relPath, matchInfo.key, line)
		if !seen[usageKey] {
			usages = append(usages, analyzer.EnvUsage{
				Key:         matchInfo.key,
				File:        relPath,
				Line:        line,
				CodeSnippet: matchInfo.codeSnippet,
				IsPartial:   matchInfo.isPartial,
				IsVarRef:    matchInfo.isVarRef,
				FullExpr:    matchInfo.fullExpr,
			})
			seen[usageKey] = true
		}
	}

	return usages, nil
}


