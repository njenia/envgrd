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
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755) // Should be excluded

	// Create files that should be scanned
	os.WriteFile(filepath.Join(tmpDir, "src", "app.js"), []byte("console.log('test');"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src", "app.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src", "app.py"), []byte("print('test')"), 0644)

	// Create file in excluded directory
	os.WriteFile(filepath.Join(tmpDir, "node_modules", "lib.js"), []byte("module.exports = {};"), 0644)

	// Create binary file (should be excluded)
	os.WriteFile(filepath.Join(tmpDir, "src", "image.png"), []byte("fake png"), 0644)

	scanner := NewScanner()
	files, err := scanner.Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should find 3 source files (js, go, py) but not the one in node_modules or the binary
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

	os.WriteFile(filepath.Join(tmpDir, "test.js"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("test"), 0644)

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

