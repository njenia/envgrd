package parser

import (
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

// LanguageLoader interface for loading language grammars
type LanguageLoader interface {
	LoadJavaScript() (*sitter.Language, error)
	LoadTypeScript() (*sitter.Language, error)
	LoadGo() (*sitter.Language, error)
	LoadPython() (*sitter.Language, error)
	LoadRust() (*sitter.Language, error)
	LoadJava() (*sitter.Language, error)
}

// DefaultLanguageLoader is a stub implementation
// This needs to be replaced with actual language grammar loading
type DefaultLanguageLoader struct{}

func (l *DefaultLanguageLoader) LoadJavaScript() (*sitter.Language, error) {
	langPtr := tree_sitter_javascript.Language()
	if langPtr == nil {
		return nil, fmt.Errorf("failed to load JavaScript language grammar")
	}
	return sitter.NewLanguage(langPtr), nil
}

func (l *DefaultLanguageLoader) LoadTypeScript() (*sitter.Language, error) {
	langPtr := tree_sitter_typescript.LanguageTypescript()
	if langPtr == nil {
		return nil, fmt.Errorf("failed to load TypeScript language grammar")
	}
	return sitter.NewLanguage(langPtr), nil
}

func (l *DefaultLanguageLoader) LoadGo() (*sitter.Language, error) {
	langPtr := tree_sitter_go.Language()
	if langPtr == nil {
		return nil, fmt.Errorf("failed to load Go language grammar")
	}
	return sitter.NewLanguage(langPtr), nil
}

func (l *DefaultLanguageLoader) LoadPython() (*sitter.Language, error) {
	langPtr := tree_sitter_python.Language()
	if langPtr == nil {
		return nil, fmt.Errorf("failed to load Python language grammar")
	}
	return sitter.NewLanguage(langPtr), nil
}

func (l *DefaultLanguageLoader) LoadRust() (*sitter.Language, error) {
	langPtr := tree_sitter_rust.Language()
	if langPtr == nil {
		return nil, fmt.Errorf("failed to load Rust language grammar")
	}
	return sitter.NewLanguage(langPtr), nil
}

func (l *DefaultLanguageLoader) LoadJava() (*sitter.Language, error) {
	langPtr := tree_sitter_java.Language()
	if langPtr == nil {
		return nil, fmt.Errorf("failed to load Java language grammar")
	}
	return sitter.NewLanguage(langPtr), nil
}

var defaultLoader LanguageLoader = &DefaultLanguageLoader{}

// SetLanguageLoader sets a custom language loader
func SetLanguageLoader(loader LanguageLoader) {
	defaultLoader = loader
}

// loadLanguage loads the Tree-Sitter language grammar for the given language
func loadLanguage(lang string) (*sitter.Language, error) {
	switch lang {
	case "javascript":
		return defaultLoader.LoadJavaScript()
	case "typescript":
		// TypeScript and TSX use the same query, but TSX files should use TSX grammar
		// For now, we'll use TypeScript grammar for both .ts and .tsx files
		// The scanner detects both as "typescript" language
		return defaultLoader.LoadTypeScript()
	case "go":
		return defaultLoader.LoadGo()
	case "python":
		return defaultLoader.LoadPython()
	case "rust":
		return defaultLoader.LoadRust()
	case "java":
		return defaultLoader.LoadJava()
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}
}

