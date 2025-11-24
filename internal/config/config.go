package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the envgrd configuration file
type Config struct {
	Ignores IgnoresConfig `yaml:"ignores"`
}

// IgnoresConfig contains ignore rules for environment variables
type IgnoresConfig struct {
	Missing []string `yaml:"missing"` // Variables to ignore when reporting as missing
	Folders []string `yaml:"folders"` // Folders to ignore when scanning (e.g., config directories)
}

// LoadConfig loads the .envgrd.config file from the specified directory
func LoadConfig(rootPath string) (*Config, error) {
	configPath := filepath.Join(rootPath, ".envgrd.config")
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file, return default config
		return &Config{
			Ignores: IgnoresConfig{
				Missing: []string{},
				Folders: []string{},
			},
		}, nil
	}
	
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return &config, nil
}

// ShouldIgnoreMissing checks if a variable should be ignored when reporting as missing
func (c *Config) ShouldIgnoreMissing(varName string) bool {
	for _, ignored := range c.Ignores.Missing {
		if ignored == varName {
			return true
		}
	}
	return false
}

// GetIgnoredMissingCount returns the number of ignored missing variables from a list
func (c *Config) GetIgnoredMissingCount(missingVars []string) int {
	count := 0
	for _, varName := range missingVars {
		if c.ShouldIgnoreMissing(varName) {
			count++
		}
	}
	return count
}

