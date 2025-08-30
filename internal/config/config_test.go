package config

import (
	"os"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	// Backup original env vars
	originalPort := os.Getenv("APP_PORT")
	originalEnv := os.Getenv("APP_ENV")
	
	// Clear env vars to test defaults
	os.Unsetenv("APP_PORT")
	os.Unsetenv("APP_ENV")
	
	defer func() {
		// Restore original env vars
		if originalPort != "" {
			os.Setenv("APP_PORT", originalPort)
		}
		if originalEnv != "" {
			os.Setenv("APP_ENV", originalEnv)
		}
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test default values
	if cfg.App.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.App.Port)
	}

	if cfg.App.Env != "development" {
		t.Errorf("Expected default environment 'development', got %s", cfg.App.Env)
	}
}

func TestConfigEnvironmentOverride(t *testing.T) {
	// Set custom env vars
	os.Setenv("APP_PORT", "9090")
	os.Setenv("APP_ENV", "test")
	
	defer func() {
		os.Unsetenv("APP_PORT")
		os.Unsetenv("APP_ENV")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test overridden values
	if cfg.App.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", cfg.App.Port)
	}

	if cfg.App.Env != "test" {
		t.Errorf("Expected environment 'test', got %s", cfg.App.Env)
	}
}
