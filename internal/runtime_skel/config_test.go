package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_NoConfigFile(t *testing.T) {
	// Use a unique app name that won't have a config file
	config, err := LoadConfig("nonexistent_app_12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Headers == nil {
		t.Error("expected Headers map to be initialized")
	}
}

func TestLoadConfig_EnvVarOverride(t *testing.T) {
	appName := "testapp"
	expectedURL := "https://api.example.com"

	// Set environment variable
	envVar := "TESTAPP_BASE_URL"
	os.Setenv(envVar, expectedURL)
	defer os.Unsetenv(envVar)

	config, err := LoadConfig(appName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.BaseURL != expectedURL {
		t.Errorf("expected BaseURL %q, got %q", expectedURL, config.BaseURL)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	// Create a temp config directory
	tmpDir := t.TempDir()
	appName := "testapp"
	configDir := filepath.Join(tmpDir, appName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config file
	configContent := `base_url: https://api.example.com
headers:
  X-Api-Key: secret123
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set XDG_CONFIG_HOME to our temp dir
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	config, err := LoadConfig(appName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.BaseURL != "https://api.example.com" {
		t.Errorf("expected BaseURL 'https://api.example.com', got %q", config.BaseURL)
	}

	if config.Headers["X-Api-Key"] != "secret123" {
		t.Errorf("expected header X-Api-Key='secret123', got %q", config.Headers["X-Api-Key"])
	}
}

func TestLoadConfig_EnvOverridesFile(t *testing.T) {
	// Create a temp config directory
	tmpDir := t.TempDir()
	appName := "testapp"
	configDir := filepath.Join(tmpDir, appName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config file with one URL
	configContent := `base_url: https://file.example.com`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set XDG_CONFIG_HOME to our temp dir
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// Set environment variable with different URL
	envURL := "https://env.example.com"
	os.Setenv("TESTAPP_BASE_URL", envURL)
	defer os.Unsetenv("TESTAPP_BASE_URL")

	config, err := LoadConfig(appName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Environment should override file
	if config.BaseURL != envURL {
		t.Errorf("expected BaseURL from env %q, got %q", envURL, config.BaseURL)
	}
}

func TestGetEnvOrConfig_EnvTakesPrecedence(t *testing.T) {
	envVar := "TEST_VAR"
	envValue := "from_env"
	os.Setenv(envVar, envValue)
	defer os.Unsetenv(envVar)

	config := &Config{
		Headers: map[string]string{
			"test": "from_config",
		},
	}

	result := GetEnvOrConfig(envVar, "test", "default", config)
	if result != envValue {
		t.Errorf("expected %q, got %q", envValue, result)
	}
}

func TestGetEnvOrConfig_ConfigUsedWhenNoEnv(t *testing.T) {
	envVar := "TEST_VAR_NOT_SET"
	os.Unsetenv(envVar)

	config := &Config{
		Headers: map[string]string{
			"test": "from_config",
		},
	}

	result := GetEnvOrConfig(envVar, "test", "default", config)
	if result != "from_config" {
		t.Errorf("expected 'from_config', got %q", result)
	}
}

func TestGetEnvOrConfig_DefaultUsedWhenNothingSet(t *testing.T) {
	envVar := "TEST_VAR_NOT_SET"
	os.Unsetenv(envVar)

	config := &Config{
		Headers: make(map[string]string),
	}

	result := GetEnvOrConfig(envVar, "nonexistent", "default_value", config)
	if result != "default_value" {
		t.Errorf("expected 'default_value', got %q", result)
	}
}

func TestGetEnvOrConfig_NilConfig(t *testing.T) {
	envVar := "TEST_VAR_NOT_SET"
	os.Unsetenv(envVar)

	result := GetEnvOrConfig(envVar, "test", "default_value", nil)
	if result != "default_value" {
		t.Errorf("expected 'default_value', got %q", result)
	}
}

func TestLoadConfig_YmlExtension(t *testing.T) {
	// Create a temp config directory
	tmpDir := t.TempDir()
	appName := "testapp"
	configDir := filepath.Join(tmpDir, appName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config file with .yml extension
	configContent := `base_url: https://yml.example.com`
	configPath := filepath.Join(configDir, "config.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set XDG_CONFIG_HOME to our temp dir
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	config, err := LoadConfig(appName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.BaseURL != "https://yml.example.com" {
		t.Errorf("expected BaseURL from .yml file, got %q", config.BaseURL)
	}
}
