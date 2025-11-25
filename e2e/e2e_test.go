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

func getSnapshotPath(testName string) string {
	return filepath.Join("testdata", "snapshots", testName+".snapshot")
}

func readSnapshot(t *testing.T, snapshotPath string) string {
	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "" // Snapshot doesn't exist yet
		}
		t.Fatalf("Failed to read snapshot: %v", err)
	}
	return string(data)
}

func writeSnapshot(t *testing.T, snapshotPath string, content string) {
	dir := filepath.Dir(snapshotPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create snapshot directory: %v", err)
	}
	if err := os.WriteFile(snapshotPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write snapshot: %v", err)
	}
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
		
		// Normalize scanning path
		if strings.HasPrefix(line, "Scanning ") {
			// Replace any temp directory paths with placeholder
			if strings.Contains(line, "/var/folders/") || strings.Contains(line, "/tmp/") {
				normalized = append(normalized, "Scanning [TEMP_DIR]...")
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

func TestE2E_BasicScan(t *testing.T) {
	mockRepo := setupMockRepo(t)
	binaryPath := getBinaryPath()
	snapshotPath := getSnapshotPath("TestE2E_BasicScan")

	// Run envgrd scan
	cmd := exec.Command(binaryPath, "scan", mockRepo)
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

	// Read existing snapshot or create new one
	expectedOutput := readSnapshot(t, snapshotPath)
	
	if expectedOutput == "" {
		// Snapshot doesn't exist - create it
		t.Logf("Creating new snapshot at %s", snapshotPath)
		writeSnapshot(t, snapshotPath, normalizedOutput)
		t.Log("Snapshot created. Run the test again to verify.")
		return
	}

	// Compare actual output with snapshot
	if normalizedOutput != expectedOutput {
		t.Errorf("Output does not match snapshot.\n\nExpected:\n%s\n\nGot:\n%s\n\nTo update the snapshot, delete %s and run the test again.", 
			expectedOutput, normalizedOutput, snapshotPath)
	}
}
