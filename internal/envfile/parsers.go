package envfile

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// detectFileType determines the type of environment file based on filename and content
func detectFileType(path string) string {
	filename := filepath.Base(path)
	
	// .envrc files (direnv)
	if filename == ".envrc" {
		return "envrc"
	}
	
	// .env.* files
	if strings.HasPrefix(filename, ".env") {
		return "env"
	}
	
	// docker-compose files
	if filename == "docker-compose.yml" || filename == "docker-compose.yaml" ||
		strings.HasPrefix(filename, "docker-compose.") {
		return "docker-compose"
	}
	
	// Kubernetes files
	if strings.HasSuffix(filename, "configmap.yaml") || strings.HasSuffix(filename, "configmap.yml") ||
		strings.HasSuffix(filename, "secret.yaml") || strings.HasSuffix(filename, "secret.yml") ||
		strings.Contains(filename, "configmap") || strings.Contains(filename, "secret") {
		// Check if it's YAML
		if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			return "k8s"
		}
	}
	
	// systemd service files
	if strings.HasSuffix(filename, ".service") {
		return "systemd"
	}
	
	// Shell scripts - check by extension or shebang
	if strings.HasSuffix(filename, ".sh") || strings.HasSuffix(filename, ".bash") {
		return "shell"
	}
	
	// Default to env format for unknown files
	return "env"
}

// parseEnvrc parses direnv .envrc files
// Supports: export VAR=value
func parseEnvrc(path string) (map[string]string, error) {
	vars := make(map[string]string)
	
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return vars, nil
		}
		return nil, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	exportRegex := regexp.MustCompile(`^\s*export\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Match export VAR=value
		matches := exportRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			value := strings.TrimSpace(matches[2])
			
			// Remove quotes
			value = trimQuotes(value)
			
			if key != "" {
				vars[key] = value
			}
		}
	}
	
	return vars, scanner.Err()
}

// parseDockerCompose parses docker-compose.yml files
func parseDockerCompose(path string) (map[string]string, error) {
	vars := make(map[string]string)
	
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return vars, nil
		}
		return nil, err
	}
	defer file.Close()
	
	var compose map[string]interface{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&compose); err != nil {
		return vars, nil // Not a valid YAML, skip silently
	}
	
	// Extract environment variables from services
	if services, ok := compose["services"].(map[string]interface{}); ok {
		for _, service := range services {
			if serviceMap, ok := service.(map[string]interface{}); ok {
				// Check environment: section
				if env, ok := serviceMap["environment"].(map[string]interface{}); ok {
					for k, v := range env {
						if val, ok := v.(string); ok {
							vars[k] = val
						} else {
							vars[k] = fmt.Sprintf("%v", v)
						}
					}
				}
				// Check environment: as array
				if envList, ok := serviceMap["environment"].([]interface{}); ok {
					for _, item := range envList {
						if envStr, ok := item.(string); ok {
							parts := strings.SplitN(envStr, "=", 2)
							if len(parts) == 2 {
								vars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
							}
						}
					}
				}
			}
		}
	}
	
	return vars, nil
}

// parseK8s parses Kubernetes ConfigMap and Secret YAML files
func parseK8s(path string) (map[string]string, error) {
	vars := make(map[string]string)
	
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return vars, nil
		}
		return nil, err
	}
	defer file.Close()
	
	var k8sObj map[string]interface{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&k8sObj); err != nil {
		return vars, nil // Not a valid YAML, skip silently
	}
	
	kind, _ := k8sObj["kind"].(string)
	
	// Handle ConfigMap
	if kind == "ConfigMap" {
		if data, ok := k8sObj["data"].(map[string]interface{}); ok {
			for k, v := range data {
				if val, ok := v.(string); ok {
					vars[k] = val
				}
			}
		}
	}
	
	// Handle Secret
	if kind == "Secret" {
		if data, ok := k8sObj["data"].(map[string]interface{}); ok {
			for k, v := range data {
				if val, ok := v.(string); ok {
					// Secrets are base64 encoded
					decoded, err := base64.StdEncoding.DecodeString(val)
					if err == nil {
						vars[k] = string(decoded)
					} else {
						vars[k] = val // Use as-is if decoding fails
					}
				}
			}
		}
	}
	
	return vars, nil
}

// parseSystemd parses systemd .service files
func parseSystemd(path string) (map[string]string, error) {
	vars := make(map[string]string)
	
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return vars, nil
		}
		return nil, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	envRegex := regexp.MustCompile(`^\s*Environment\s*=\s*(.+)$`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Match Environment=VAR=value or Environment="VAR=value"
		matches := envRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			envStr := strings.TrimSpace(matches[1])
			// Remove quotes if present
			envStr = trimQuotes(envStr)
			parts := strings.SplitN(envStr, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if key != "" {
					vars[key] = value
				}
			}
		}
	}
	
	return vars, scanner.Err()
}

// parseShellScript parses shell scripts for export VAR=value
func parseShellScript(path string) (map[string]string, error) {
	vars := make(map[string]string)
	
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return vars, nil
		}
		return nil, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	exportRegex := regexp.MustCompile(`^\s*export\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Match export VAR=value
		matches := exportRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			value := strings.TrimSpace(matches[2])
			
			// Remove quotes
			value = trimQuotes(value)
			
			if key != "" {
				vars[key] = value
			}
		}
	}
	
	return vars, scanner.Err()
}

// trimQuotes removes surrounding quotes from a string
func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') ||
			(s[0] == '`' && s[len(s)-1] == '`') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

