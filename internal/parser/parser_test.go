package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParser_JavaScript_StaticPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.js")
	
	code := `
const apiKey = process.env.API_KEY;
const dbUrl = process.env["DATABASE_URL"];
const secret = process.env['SECRET_KEY'];
`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "javascript", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	expected := []string{"API_KEY", "DATABASE_URL", "SECRET_KEY"}
	if len(usages) != len(expected) {
		t.Errorf("Expected %d usages, got %d", len(expected), len(usages))
	}
	
	keys := make(map[string]bool)
	for _, usage := range usages {
		keys[usage.Key] = true
		if usage.IsPartial {
			t.Errorf("Expected static match, got partial for key: %s", usage.Key)
		}
	}
	
	for _, key := range expected {
		if !keys[key] {
			t.Errorf("Missing expected key: %s", key)
		}
	}
}

func TestParser_JavaScript_DynamicPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.js")
	
	code := `
const prefix = "API_";
const key1 = process.env[prefix + "KEY"];
const key2 = process.env["PREFIX_" + suffix];
const key3 = process.env[varName];
`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "javascript", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	// Should find 3 partial matches
	if len(usages) < 3 {
		t.Errorf("Expected at least 3 usages, got %d", len(usages))
	}
	
	partialCount := 0
	varRefCount := 0
	for _, usage := range usages {
		if usage.IsPartial {
			partialCount++
		}
		if usage.IsVarRef {
			varRefCount++
		}
	}
	
	if partialCount < 2 {
		t.Errorf("Expected at least 2 partial matches, got %d", partialCount)
	}
	if varRefCount < 1 {
		t.Errorf("Expected at least 1 variable reference, got %d", varRefCount)
	}
}

func TestParser_TypeScript_StaticPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.ts")
	
	code := `
const apiKey: string = process.env.API_KEY || "";
const dbUrl = process.env["DATABASE_URL"];
`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "typescript", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	expected := []string{"API_KEY", "DATABASE_URL"}
	keys := make(map[string]bool)
	for _, usage := range usages {
		keys[usage.Key] = true
		if usage.IsPartial {
			t.Errorf("Expected static match, got partial for key: %s", usage.Key)
		}
	}
	
	for _, key := range expected {
		if !keys[key] {
			t.Errorf("Missing expected key: %s", key)
		}
	}
}

func TestParser_Go_StaticPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.go")
	
	code := `
package main

import "os"

func main() {
	apiKey := os.Getenv("API_KEY")
	dbUrl := os.Getenv("DATABASE_URL")
}
`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "go", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	expected := []string{"API_KEY", "DATABASE_URL"}
	keys := make(map[string]bool)
	for _, usage := range usages {
		keys[usage.Key] = true
		if usage.IsPartial {
			t.Errorf("Expected static match, got partial for key: %s", usage.Key)
		}
	}
	
	for _, key := range expected {
		if !keys[key] {
			t.Errorf("Missing expected key: %s", key)
		}
	}
}

func TestParser_Go_DynamicPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.go")
	
	code := `
package main

import "os"

func main() {
	prefix := "API_"
	key1 := os.Getenv(prefix + "KEY")
	key2 := os.Getenv("PREFIX_" + suffix)
	key3 := os.Getenv(varName)
}
`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "go", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	// Should find 3 partial matches
	if len(usages) < 3 {
		t.Errorf("Expected at least 3 usages, got %d", len(usages))
	}
	
	partialCount := 0
	varRefCount := 0
	for _, usage := range usages {
		if usage.IsPartial {
			partialCount++
		}
		if usage.IsVarRef {
			varRefCount++
		}
	}
	
	if partialCount < 2 {
		t.Errorf("Expected at least 2 partial matches, got %d", partialCount)
	}
	if varRefCount < 1 {
		t.Errorf("Expected at least 1 variable reference, got %d", varRefCount)
	}
}

func TestParser_Python_StaticPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.py")
	
	code := `
import os

api_key = os.environ["API_KEY"]
db_url = os.getenv("DATABASE_URL")
secret = os.environ['SECRET_KEY']
`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "python", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	expected := []string{"API_KEY", "DATABASE_URL", "SECRET_KEY"}
	keys := make(map[string]bool)
	for _, usage := range usages {
		keys[usage.Key] = true
		if usage.IsPartial {
			t.Errorf("Expected static match, got partial for key: %s", usage.Key)
		}
	}
	
	for _, key := range expected {
		if !keys[key] {
			t.Errorf("Missing expected key: %s", key)
		}
	}
}

func TestParser_Python_DynamicPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.py")
	
	code := `
import os

prefix = "API_"
key1 = os.environ[prefix + "KEY"]
key2 = os.getenv("PREFIX_" + suffix)
key3 = os.environ[varName]
key4 = os.getenv(varName)
`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "python", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	// Should find at least 4 partial matches
	if len(usages) < 4 {
		t.Errorf("Expected at least 4 usages, got %d", len(usages))
	}
	
	partialCount := 0
	varRefCount := 0
	for _, usage := range usages {
		if usage.IsPartial {
			partialCount++
		}
		if usage.IsVarRef {
			varRefCount++
		}
	}
	
	if partialCount < 2 {
		t.Errorf("Expected at least 2 partial matches, got %d", partialCount)
	}
	if varRefCount < 2 {
		t.Errorf("Expected at least 2 variable references, got %d", varRefCount)
	}
}

func TestParser_Rust_StaticPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.rs")
	
	code := `
use std::env;

fn main() {
	let api_key = env::var("API_KEY").unwrap();
	let db_url = std::env::var("DATABASE_URL").unwrap();
	let secret = env::var_os("SECRET_KEY");
}
`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "rust", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	expected := []string{"API_KEY", "DATABASE_URL", "SECRET_KEY"}
	keys := make(map[string]bool)
	for _, usage := range usages {
		keys[usage.Key] = true
		if usage.IsPartial {
			t.Errorf("Expected static match, got partial for key: %s", usage.Key)
		}
	}
	
	for _, key := range expected {
		if !keys[key] {
			t.Errorf("Missing expected key: %s", key)
		}
	}
}

func TestParser_Rust_DynamicPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.rs")
	
	code := `
use std::env;

fn main() {
	let prefix = "API_";
	let key1 = env::var(prefix.to_string() + "KEY").unwrap();
	let key2 = std::env::var("PREFIX_".to_string() + &suffix).unwrap();
	let key3 = env::var(var_name).unwrap();
}
`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "rust", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	// Should find at least 3 partial matches
	if len(usages) < 3 {
		t.Errorf("Expected at least 3 usages, got %d", len(usages))
	}
	
	partialCount := 0
	varRefCount := 0
	for _, usage := range usages {
		if usage.IsPartial {
			partialCount++
		}
		if usage.IsVarRef {
			varRefCount++
		}
	}
	
	if partialCount < 2 {
		t.Errorf("Expected at least 2 partial matches, got %d", partialCount)
	}
	if varRefCount < 1 {
		t.Errorf("Expected at least 1 variable reference, got %d", varRefCount)
	}
}

func TestParser_Java_StaticPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "Test.java")
	
	code := `
public class Test {
	public static void main(String[] args) {
		String apiKey = System.getenv("API_KEY");
		String dbUrl = System.getenv().get("DATABASE_URL");
	}
}
`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "java", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	expected := []string{"API_KEY", "DATABASE_URL"}
	keys := make(map[string]bool)
	for _, usage := range usages {
		keys[usage.Key] = true
		if usage.IsPartial {
			t.Errorf("Expected static match, got partial for key: %s", usage.Key)
		}
	}
	
	for _, key := range expected {
		if !keys[key] {
			t.Errorf("Missing expected key: %s", key)
		}
	}
}

func TestParser_Java_DynamicPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "Test.java")
	
	code := `
public class Test {
	public static void main(String[] args) {
		String prefix = "API_";
		String key1 = System.getenv(prefix + "KEY");
		String key2 = System.getenv().get("PREFIX_" + suffix);
		String key3 = System.getenv(varName);
	}
}
`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "java", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	// Should find at least 3 partial matches
	if len(usages) < 3 {
		t.Errorf("Expected at least 3 usages, got %d", len(usages))
	}
	
	partialCount := 0
	varRefCount := 0
	for _, usage := range usages {
		if usage.IsPartial {
			partialCount++
		}
		if usage.IsVarRef {
			varRefCount++
		}
	}
	
	if partialCount < 2 {
		t.Errorf("Expected at least 2 partial matches, got %d", partialCount)
	}
	if varRefCount < 1 {
		t.Errorf("Expected at least 1 variable reference, got %d", varRefCount)
	}
}

func TestParser_LineNumbers(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.js")
	
	code := `const key1 = process.env.KEY1;
const key2 = process.env.KEY2;
const key3 = process.env.KEY3;
`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "javascript", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	if len(usages) != 3 {
		t.Fatalf("Expected 3 usages, got %d", len(usages))
	}
	
	// Check line numbers
	expectedLines := []int{1, 2, 3}
	actualLines := make([]int, len(usages))
	for i, usage := range usages {
		actualLines[i] = usage.Line
	}
	
	if !reflect.DeepEqual(actualLines, expectedLines) {
		t.Errorf("Expected lines %v, got %v", expectedLines, actualLines)
	}
}

func TestParser_CodeSnippets(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.js")
	
	code := `const apiKey = process.env.API_KEY;`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "javascript", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	if len(usages) != 1 {
		t.Fatalf("Expected 1 usage, got %d", len(usages))
	}
	
	if usages[0].CodeSnippet == "" {
		t.Error("Expected code snippet to be populated")
	}
	
	if !contains(usages[0].CodeSnippet, "process.env.API_KEY") {
		t.Errorf("Code snippet should contain 'process.env.API_KEY', got: %s", usages[0].CodeSnippet)
	}
}

func TestParser_RelativePaths(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	filePath := filepath.Join(subDir, "test.js")
	code := `const key = process.env.KEY;`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	usages, err := parser.ParseFile(filePath, "javascript", tmpDir)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	if len(usages) != 1 {
		t.Fatalf("Expected 1 usage, got %d", len(usages))
	}
	
	// Should be relative to scan root
	expectedPath := "src/test.js"
	if usages[0].File != expectedPath {
		t.Errorf("Expected relative path %s, got %s", expectedPath, usages[0].File)
	}
}

func TestParser_InvalidLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.js")
	code := `const key = process.env.KEY;`
	
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	parser := NewParser()
	_, err := parser.ParseFile(filePath, "invalid_lang", tmpDir)
	if err == nil {
		t.Error("Expected error for invalid language, got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

