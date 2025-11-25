package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
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

func setupMockRepo(t *testing.T, repoName string) string {
	// Get the testdata directory
	testdataDir := filepath.Join("testdata", repoName)

	// Check if testdata directory exists
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Fatalf("Testdata directory not found: %s", testdataDir)
	}

	// Get absolute path to testdata directory
	absPath, err := filepath.Abs(testdataDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// envgrd scan is read-only, so we can use testdata directly
	return absPath
}

func normalizeOutput(output string) string {
	// Normalize output for consistent comparison
	// Remove ANSI color codes
	output = removeANSICodes(output)

	// Remove any paths that might vary (like temp directories)
	lines := strings.Split(output, "\n")
	var normalized []string
	for _, line := range lines {
		// Normalize version line (version will vary)
		if strings.HasPrefix(line, "Version: ") {
			normalized = append(normalized, "Version: [VERSION]")
			continue
		}

		// Normalize scanning path (replace any absolute path with placeholder)
		if strings.HasPrefix(line, "Scanning ") {
			// Extract the path part and replace with placeholder
			if idx := strings.Index(line, "..."); idx > 0 {
				normalized = append(normalized, "Scanning [SCAN_DIR]...")
			} else {
				normalized = append(normalized, line)
			}
		} else {
			normalized = append(normalized, line)
		}
	}
	return strings.Join(normalized, "\n")
}

func removeANSICodes(s string) string {
	// Remove ANSI escape sequences (e.g., [1m, [33m, [0m, [90m)
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' || s[i] == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}

func runScanTest(t *testing.T, repoName string, envVars map[string]string) {
	mockRepo := setupMockRepo(t, repoName)
	binaryPath := getBinaryPath()

	// Run envgrd scan
	cmd := exec.Command(binaryPath, "scan", mockRepo)

	// Set environment variables if provided
	if envVars != nil {
		cmd.Env = os.Environ()
		for k, v := range envVars {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	output, err := cmd.CombinedOutput()

	outputStr := string(output)
	normalizedOutput := normalizeOutput(outputStr)

	// Handle exit code (exit code 1 is expected when there are unused variables)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 1 {
				t.Fatalf("Unexpected exit code: %d\nOutput: %s", exitError.ExitCode(), outputStr)
			}
			// Exit code 1 is expected when unused variables are found
		} else {
			t.Fatalf("envgrd scan failed: %v\nOutput: %s", err, outputStr)
		}
	}

	// Use cupaloy for snapshot testing
	cupaloy.SnapshotT(t, normalizedOutput)
}

func TestE2E_BasicScan(t *testing.T) {
	runScanTest(t, "mock-repo", nil)
}

func TestE2E_ConfigIgnores(t *testing.T) {
	// Test that variables in ignores.missing are not reported as missing
	// and that files in ignored folders are not scanned
	runScanTest(t, "mock-repo-ignores", nil)
}

func TestE2E_MultiLanguage(t *testing.T) {
	// Test scanning files in multiple languages (Python, Rust, Java, TypeScript)
	runScanTest(t, "mock-repo-multilang", nil)
}

func TestE2E_MultipleEnvFiles(t *testing.T) {
	// Test that envgrd detects and loads multiple env files (.env, .env.production, .env.local)
	runScanTest(t, "mock-repo-envfiles", nil)
}

func TestE2E_ExportedVars(t *testing.T) {
	// Test that exported environment variables are recognized and prevent false positives
	envVars := map[string]string{
		"CI_TOKEN": "ci-token-value",
	}
	runScanTest(t, "mock-repo-exported", envVars)
}
