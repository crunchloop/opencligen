package runtime

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultWarningWriter is the writer used for security warnings
var DefaultWarningWriter io.Writer = os.Stderr

// Config holds the CLI configuration
type Config struct {
	BaseURL string            `yaml:"base_url"`
	Headers map[string]string `yaml:"headers"`
}

// LoadConfig loads configuration from file and environment
func LoadConfig(appName string) (*Config, error) {
	config := &Config{
		Headers: make(map[string]string),
	}

	// Try to load from config file
	configPath := getConfigPath(appName)
	if configPath != "" {
		if err := loadConfigFile(configPath, config); err != nil {
			// Config file is optional, ignore errors
			_ = err
		}
	}

	// Environment variables override config file
	envPrefix := strings.ToUpper(appName) + "_"
	if baseURL := os.Getenv(envPrefix + "BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}

	return config, nil
}

// getConfigPath returns the path to the config file
func getConfigPath(appName string) string {
	// Try XDG_CONFIG_HOME first
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configHome = filepath.Join(home, ".config")
	}

	configPath := filepath.Join(configHome, appName, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Try .yml extension
	configPath = filepath.Join(configHome, appName, "config.yml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	return ""
}

// loadConfigFile loads configuration from a YAML file
func loadConfigFile(path string, config *Config) error {
	// Check file permissions for security
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// Warn if config file is readable by others (potentially contains secrets)
	mode := info.Mode().Perm()
	if mode&0044 != 0 { // Check if group or others have read permission
		fmt.Fprintf(DefaultWarningWriter, "Warning: config file %s has insecure permissions %o. "+
			"Consider running: chmod 600 %s\n", path, mode, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}

// GetEnvOrConfig returns a value from environment, config, or default
func GetEnvOrConfig(envVar, configKey, defaultValue string, config *Config) string {
	// Environment takes precedence
	if val := os.Getenv(envVar); val != "" {
		return val
	}

	// Then config
	if config != nil && configKey != "" {
		// For now, we only support headers in config
		if val, ok := config.Headers[configKey]; ok {
			return val
		}
	}

	return defaultValue
}
