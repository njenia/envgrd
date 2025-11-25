package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jenian/envgrd/internal/analyzer"
	"github.com/jenian/envgrd/internal/config"
	"github.com/jenian/envgrd/internal/envfile"
	"github.com/jenian/envgrd/internal/output"
	"github.com/jenian/envgrd/internal/parser"
	"github.com/jenian/envgrd/internal/scanner"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags
var Version = "dev"

// envVarData holds processed environment variable data
type envVarData struct {
	envVars              map[string]string // All env vars (from files + exported)
	envVarsFromFilesOnly map[string]string // Only vars from .env files (for unused check)
	relEnvKeySources     map[string]string // Relative paths to source files
}

var (
	rootCmd = &cobra.Command{
		Use:   "envgrd",
		Short: "Scan codebase for environment variable usages",
		Long:  "A CLI tool that scans codebases for environment variable usages and compares them with .env files.",
	}

	scanCmd = &cobra.Command{
		Use:   "scan [path]",
		Short: "Scan a codebase for environment variable usages",
		Long:  "Recursively scan a directory for environment variable usages and compare with .env files.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runScan,
	}

	initSchemaCmd = &cobra.Command{
		Use:   "init-schema",
		Short: "Generate a schema template (stub for future feature)",
		Long:  "Generate a JSON schema template for environment variable validation.",
		RunE:  runInitSchema,
	}

	initConfigCmd = &cobra.Command{
		Use:   "init-config",
		Short: "Create a .envgrd.config file in the current directory",
		Long:  "Creates a .envgrd.config file with default configuration in the current directory.",
		RunE:  runInitConfig,
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long:  "Print the version number of envgrd",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}

	// Flags
	scanPath     string
	envFile      string
	jsonOutput   bool
	silent       bool
	skipUnused   bool
	debug        bool
	noHeader     bool
	noDynamic    bool
	includeGlobs []string
	excludeGlobs []string
)

func init() {
	scanCmd.Flags().StringVarP(&scanPath, "path", "p", ".", "Path to scan (default: current directory)")
	scanCmd.Flags().StringVar(&envFile, "env-file", "", "Additional .env file to load")
	scanCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	scanCmd.Flags().BoolVar(&silent, "silent", false, "Silent mode (exit code only)")
	scanCmd.Flags().BoolVar(&skipUnused, "skip-unused", false, "Skip reporting unused variables")
	scanCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")
	scanCmd.Flags().BoolVar(&noHeader, "no-header", false, "Skip printing the header")
	scanCmd.Flags().BoolVar(&noDynamic, "no-dynamic", false, "Disable dynamic pattern detection (skip partial matches from runtime-evaluated expressions)")
	scanCmd.Flags().StringSliceVar(&includeGlobs, "include", []string{}, "Glob patterns to include")
	scanCmd.Flags().StringSliceVar(&excludeGlobs, "exclude", []string{}, "Glob patterns to exclude")

	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(initSchemaCmd)
	rootCmd.AddCommand(initConfigCmd)
	rootCmd.AddCommand(versionCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	// Get scan path
	path := scanPath
	if len(args) > 0 {
		path = args[0]
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	fileScanner := scanner.NewScanner()
	if len(includeGlobs) > 0 {
		fileScanner.SetIncludeGlobs(includeGlobs)
	}
	if len(excludeGlobs) > 0 {
		fileScanner.SetExcludeGlobs(excludeGlobs)
	}

	envLoader := envfile.NewLoader()
	if envFile != "" {
		envLoader.AddEnvFile(envFile)
	}

	tsParser := parser.NewParser()
	tsParser.SetDebug(debug)

	// Print header unless disabled or in JSON/silent mode
	if !noHeader && !jsonOutput && !silent {
		printHeader()
	}

	cfg, err := config.LoadConfig(absPath)
	if err != nil {
		if !silent {
			fmt.Fprintf(os.Stderr, "Warning: failed to load .envgrd.config: %v\n", err)
		}
		// Continue with default config
		cfg = &config.Config{}
	}

	if len(cfg.Ignores.Folders) > 0 {
		fileScanner.AddExcludeDirs(cfg.Ignores.Folders)
	}

	if !silent {
		fmt.Fprintf(os.Stderr, "Scanning %s...\n", absPath)
	}
	files, err := fileScanner.Scan(absPath)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if !silent {
		report := reportFileCounts(files)
		fmt.Fprintf(os.Stderr, "%s\n", report)
	}

	envData, err := loadEnvironmentVariables(envLoader, absPath)
	if err != nil {
		return err
	}

	allUsages := parseFiles(tsParser, files, absPath, silent)

	result := analyzer.Analyze(allUsages, envData.envVars, envData.envVarsFromFilesOnly, envData.relEnvKeySources, cfg)

	dynamic := !noDynamic
	if err := output.Format(result, jsonOutput, silent, skipUnused, dynamic); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	if output.HasIssues(result, skipUnused, dynamic) {
		os.Exit(1)
	}

	return nil
}

// reportFileCounts generates a formatted report string of file counts by language
func reportFileCounts(files []scanner.FileInfo) string {
	// Count files by language
	langCounts := make(map[string]int)
	for _, file := range files {
		lang := string(file.Language)
		if lang == "" {
			lang = "unknown"
		}
		langCounts[lang]++
	}

	// Build report string
	var reportParts []string
	langOrder := []string{"javascript", "typescript", "go", "python", "rust", "java"}
	for _, lang := range langOrder {
		if count, ok := langCounts[lang]; ok && count > 0 {
			// Use short names for display
			shortName := lang
			switch lang {
			case "javascript":
				shortName = "js"
			case "typescript":
				shortName = "ts"
			}
			reportParts = append(reportParts, fmt.Sprintf("%s: %d", shortName, count))
			delete(langCounts, lang)
		}
	}
	// Add any remaining languages
	for lang, count := range langCounts {
		if count > 0 {
			reportParts = append(reportParts, fmt.Sprintf("%s: %d", lang, count))
		}
	}

	if len(reportParts) > 0 {
		reportStr := ""
		for i, part := range reportParts {
			if i > 0 {
				reportStr += ", "
			}
			reportStr += part
		}
		return fmt.Sprintf("Found %d files (%s)", len(files), reportStr)
	}
	return fmt.Sprintf("Found %d files to parse", len(files))
}

// loadEnvironmentVariables loads and processes environment variables from files and exported env
func loadEnvironmentVariables(envLoader *envfile.Loader, absPath string) (*envVarData, error) {
	// Load environment variables from files and merge with exported env
	envVars, envVarsFromFilesOnly, envKeySources, err := envLoader.LoadWithExportedEnv(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load env files: %w", err)
	}

	// Make source file paths relative to scan root for better display
	relEnvKeySources := make(map[string]string)
	for k, sourcePath := range envKeySources {
		if rel, err := filepath.Rel(absPath, sourcePath); err == nil && rel != "" {
			relEnvKeySources[k] = rel
		} else {
			// Fallback to just the filename if relative path fails
			relEnvKeySources[k] = filepath.Base(sourcePath)
		}
	}

	return &envVarData{
		envVars:              envVars,
		envVarsFromFilesOnly: envVarsFromFilesOnly,
		relEnvKeySources:     relEnvKeySources,
	}, nil
}

// parses all files in parallel and returns environment variable usages
func parseFiles(tsParser *parser.Parser, files []scanner.FileInfo, absPath string, silent bool) []analyzer.EnvUsage {
	var allUsages []analyzer.EnvUsage
	var wg sync.WaitGroup
	var mu sync.Mutex
	workers := make(chan struct{}, 10)

	for _, file := range files {
		wg.Add(1)
		workers <- struct{}{} // Acquire worker

		go func(f scanner.FileInfo) {
			defer wg.Done()
			defer func() { <-workers }() // Release worker

			usages, err := tsParser.ParseFile(f.Path, string(f.Language), absPath)
			if err != nil {
				// Log error but continue
				if !silent {
					fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", f.Path, err)
				}
				return
			}

			// Mark usages from ignored folders
			if f.InIgnoredPath {
				for i := range usages {
					usages[i].InIgnoredPath = true
				}
			}

			mu.Lock()
			allUsages = append(allUsages, usages...)
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	return allUsages
}

func runInitSchema(cmd *cobra.Command, args []string) error {
	// Stub for future schema feature
	schema := `{
  "PORT": "number",
  "LOG_LEVEL": ["debug", "info", "warn", "error"]
}`
	fmt.Println(schema)
	return nil
}

func runInitConfig(cmd *cobra.Command, args []string) error {
	configPath := ".envgrd.config"

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf(".envgrd.config already exists in the current directory")
	}

	// Default config content
	configContent := `# .envgrd.config
# Configuration file for envgrd

ignores:
  # Variables that are configured in custom ways (not in .env files or standard configs)
  # These will not be reported as missing
  missing:
    # - CUSTOM_API_KEY
    # - EXTERNAL_SERVICE_TOKEN
    # Add more variable names here as needed
  
  # Folders to ignore when scanning (useful for config directories that aren't actual code)
  folders:
    # - config
    # - configs
    # - k8s
    # - deployments
    # Add more folder names here as needed
`

	// Write the config file
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create .envgrd.config: %w", err)
	}

	fmt.Printf("Created .envgrd.config in the current directory\n")
	return nil
}

func printHeader() {
	header := `  ____ __  __ __ __   ___  ____  ____  
 ||    ||\ || || ||  // \\ || \\ || \\ 
 ||==  ||\\|| \\ // (( ___ ||_// ||  ))
 ||___ || \||  \V/   \\_|| || \\ ||_// 
                                                          
`
	fmt.Print(header)
	fmt.Printf("Version: %s\n\n", Version)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
