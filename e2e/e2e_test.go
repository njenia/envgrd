package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary for testing
	buildDir := filepath.Join("..", "bin")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		panic("Failed to create bin directory: " + err.Error())
	}

	binaryPath = filepath.Join(buildDir, "envgrd")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../cmd/envgrd")
	buildCmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	if err := buildCmd.Run(); err != nil {
		panic("Failed to build binary: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.Exit(code)
}

func setupMockRepo(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create source files with environment variable usage
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}

	// Create Go file with env var usage
	goFile := filepath.Join(srcDir, "main.go")
	goCode := `package main

import (
	"fmt"
	"os"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	dbUrl := os.Getenv("DATABASE_URL")
	fmt.Println(apiKey, dbUrl)
}
`
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		t.Fatalf("Failed to write Go file: %v", err)
	}

	// Create JavaScript file with env var usage
	jsFile := filepath.Join(srcDir, "config.js")
	jsCode := `const apiKey = process.env.API_KEY;
const dbUrl = process.env["DATABASE_URL"];
const secret = process.env.SECRET_KEY;
`
	if err := os.WriteFile(jsFile, []byte(jsCode), 0644); err != nil {
		t.Fatalf("Failed to write JS file: %v", err)
	}

	// Create .env file with some variables
	envFile := filepath.Join(tmpDir, ".env")
	envContent := `API_KEY=test123
DATABASE_URL=postgres://localhost/db
SECRET_KEY=secret123
UNUSED_VAR=unused
`
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	return tmpDir
}

func TestE2E_BasicScan(t *testing.T) {
	mockRepo := setupMockRepo(t)

	// Run envgrd scan
	cmd := exec.Command(binaryPath, "scan", mockRepo)
	output, err := cmd.CombinedOutput()
	
	outputStr := string(output)

	// Verify that the scan found files
	if !strings.Contains(outputStr, "Found") {
		t.Errorf("Expected output to contain 'Found', got: %s", outputStr)
	}

	// Verify that it found the expected files (js and go)
	if !strings.Contains(outputStr, "js: 1") || !strings.Contains(outputStr, "go: 1") {
		t.Errorf("Expected output to contain file counts, got: %s", outputStr)
	}

	// Verify that unused variables are detected (UNUSED_VAR should be reported)
	if !strings.Contains(outputStr, "UNUSED_VAR") {
		t.Errorf("Expected output to contain 'UNUSED_VAR', got: %s", outputStr)
	}

	// Verify that all used variables are found (no missing variables)
	// API_KEY, DATABASE_URL, SECRET_KEY should all be in .env
	if strings.Contains(outputStr, "Missing variables:") {
		t.Errorf("Expected no missing variables, but found some in output: %s", outputStr)
	}

	// Exit code 1 is expected when there are unused variables (this is correct behavior)
	// We verify the scan worked correctly by checking the output content
	if err != nil {
		// Check if it's just an exit code error (which is expected for unused vars)
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 1 {
				t.Fatalf("Unexpected exit code: %d\nOutput: %s", exitError.ExitCode(), outputStr)
			}
			// Exit code 1 is expected when unused variables are found
		} else {
			t.Fatalf("envgrd scan failed: %v\nOutput: %s", err, outputStr)
		}
	}
}

