package envfile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Loader handles loading and parsing environment files
type Loader struct {
	envFiles []string
	autoDetect bool
}

// NewLoader creates a new env file loader
func NewLoader() *Loader {
	return &Loader{
		envFiles: []string{".env", ".env.local", "env.example"},
		autoDetect: true,
	}
}

// SetAutoDetect enables or disables automatic detection of env files
func (l *Loader) SetAutoDetect(enabled bool) {
	l.autoDetect = enabled
}

// AddEnvFile adds a custom env file to load
func (l *Loader) AddEnvFile(path string) {
	l.envFiles = append(l.envFiles, path)
}

// SetEnvFiles sets the list of env files to load
func (l *Loader) SetEnvFiles(files []string) {
	l.envFiles = files
}

// parseEnvFile parses a single environment file using the appropriate parser
func parseEnvFile(path string) (map[string]string, error) {
	fileType := detectFileType(path)
	
	switch fileType {
	case "envrc":
		return parseEnvrc(path)
	case "docker-compose":
		return parseDockerCompose(path)
	case "k8s":
		return parseK8s(path)
	case "systemd":
		return parseSystemd(path)
	case "shell":
		return parseShellScript(path)
	case "env":
		fallthrough
	default:
		return parseDotEnv(path)
	}
}

// parseDotEnv parses a standard .env file
func parseDotEnv(path string) (map[string]string, error) {
	vars := make(map[string]string)

	file, err := os.Open(path)
	if err != nil {
		// File doesn't exist, return empty map (not an error)
		if os.IsNotExist(err) {
			return vars, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Skip malformed lines (could be multiline values, etc.)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = trimQuotes(value)

		if key != "" {
			vars[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", path, err)
	}

	return vars, nil
}

// findEnvFiles finds all environment variable files in the directory
func (l *Loader) findEnvFiles(rootPath string) ([]string, error) {
	var files []string
	
	// Add explicitly configured files
	for _, envFile := range l.envFiles {
		var path string
		if filepath.IsAbs(envFile) {
			path = envFile
		} else {
			path = filepath.Join(rootPath, envFile)
		}
		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		}
	}
	
	// Auto-detect additional files if enabled
	if l.autoDetect {
		entries, err := os.ReadDir(rootPath)
		if err != nil {
			return files, nil // Can't read directory, return what we have
		}
		
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			
			name := entry.Name()
			filePath := filepath.Join(rootPath, name)
			
			// Check if it's an env file we should parse
			fileType := detectFileType(filePath)
			shouldInclude := false
			
			switch fileType {
			case "envrc":
				shouldInclude = true
			case "env":
				// Include .env.* files (but not ones already in default list)
				if strings.HasPrefix(name, ".env") {
					// Skip if already in default list
					alreadyInDefault := false
					for _, defaultFile := range []string{".env", ".env.local", "env.example"} {
						if name == defaultFile {
							alreadyInDefault = true
							break
						}
					}
					if !alreadyInDefault {
						shouldInclude = true
					}
				}
			case "docker-compose":
				shouldInclude = true
			case "k8s":
				shouldInclude = true
			case "systemd":
				shouldInclude = true
			case "shell":
				// Include .sh and .bash files
				if strings.HasSuffix(name, ".sh") || strings.HasSuffix(name, ".bash") {
					shouldInclude = true
				}
			}
			
			if shouldInclude {
				// Check if already in list
				alreadyIncluded := false
				for _, existing := range files {
					if existing == filePath {
						alreadyIncluded = true
						break
					}
				}
				if !alreadyIncluded {
					files = append(files, filePath)
				}
			}
		}
	}
	
	return files, nil
}

// Load loads all configured env files and merges them
// Later files override earlier ones
func (l *Loader) Load(rootPath string) (map[string]string, error) {
	allVars := make(map[string]string)

	// Find all env files (explicit + auto-detected)
	envFiles, err := l.findEnvFiles(rootPath)
	if err != nil {
		return nil, err
	}

	for _, path := range envFiles {
		vars, err := parseEnvFile(path)
		if err != nil {
			// Log error but continue with other files
			continue
		}

		// Merge: later files override earlier ones
		for k, v := range vars {
			allVars[k] = v
		}
	}

	return allVars, nil
}

// LoadFromPath loads env files from a specific directory
func (l *Loader) LoadFromPath(dirPath string) (map[string]string, error) {
	return l.Load(dirPath)
}

