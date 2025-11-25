package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path     string
		expected Language
	}{
		{"test.js", LanguageJavaScript},
		{"test.jsx", LanguageJavaScript},
		{"test.mjs", LanguageJavaScript},
		{"test.ts", LanguageTypeScript},
		{"test.tsx", LanguageTypeScript},
		{"test.go", LanguageGo},
		{"test.py", LanguagePython},
		{"test.txt", LanguageUnknown},
		{"test", LanguageUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectLanguage(tt.path)
			if result != tt.expected {
				t.Errorf("detectLanguage(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestScanner_Scan(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	if err := os.MkdirAll(filepath.Join(tmpDir, "src"), 0755); err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755); err != nil {
		t.Fatalf("Failed to create node_modules directory: %v", err)
	}

	// Create files that should be scanned
	if err := os.WriteFile(filepath.Join(tmpDir, "src", "app.js"), []byte("console.log('test');"), 0644); err != nil {
		t.Fatalf("Failed to write app.js: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "src", "app.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to write app.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "src", "app.py"), []byte("print('test')"), 0644); err != nil {
		t.Fatalf("Failed to write app.py: %v", err)
	}

	// Create file in excluded directory
	if err := os.WriteFile(filepath.Join(tmpDir, "node_modules", "lib.js"), []byte("module.exports = {};"), 0644); err != nil {
		t.Fatalf("Failed to write lib.js: %v", err)
	}

	// Create file with unsupported extension (should be excluded by whitelist)
	if err := os.WriteFile(filepath.Join(tmpDir, "src", "readme.txt"), []byte("readme content"), 0644); err != nil {
		t.Fatalf("Failed to write readme.txt: %v", err)
	}

	scanner := NewScanner()
	files, err := scanner.Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should find 3 source files (js, go, py) but not the one in node_modules or unsupported extensions
	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}

	// Check that node_modules is excluded
	for _, file := range files {
		if filepath.Base(filepath.Dir(file.Path)) == "node_modules" {
			t.Error("Files in node_modules should be excluded")
		}
	}
}

func TestScanner_ExcludeGlobs(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "test.js"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test.js: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test.go: %v", err)
	}

	scanner := NewScanner()
	scanner.SetExcludeGlobs([]string{"*.go"})

	files, err := scanner.Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only find .js file
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if files[0].Language != LanguageJavaScript {
		t.Errorf("Expected JavaScript file, got %v", files[0].Language)
	}
}

