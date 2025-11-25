package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func getBinaryPath() string {
	// Try ./envgrd first
	if _, err := os.Stat("./envgrd"); err == nil {
		return "./envgrd"
	}
	// Try bin/envgrd
	if _, err := os.Stat("bin/envgrd"); err == nil {
		return "bin/envgrd"
	}
	// Fallback to just "envgrd" (assumes it's in PATH)
	return "envgrd"
}

func setupMockRepo(t *testing.T) string {
	// Get the testdata directory
	testdataDir := filepath.Join("testdata", "mock-repo")
	
	// Check if testdata directory exists
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Fatalf("Testdata directory not found: %s", testdataDir)
	}

	// Copy testdata to a temporary directory
	tmpDir := t.TempDir()
	
	// Copy all files from testdata to tmpDir
	err := filepath.Walk(testdataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Get relative path from testdataDir
		relPath, err := filepath.Rel(testdataDir, path)
		if err != nil {
			return err
		}
		
		destPath := filepath.Join(tmpDir, relPath)
		
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}
		
		// Read source file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		
		// Write to destination
		return os.WriteFile(destPath, data, info.Mode())
	})
	
	if err != nil {
		t.Fatalf("Failed to copy testdata: %v", err)
	}
	
	return tmpDir
}

func TestE2E_BasicScan(t *testing.T) {
	mockRepo := setupMockRepo(t)
	binaryPath := getBinaryPath()

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
