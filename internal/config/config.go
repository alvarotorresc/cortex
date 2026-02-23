package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all runtime configuration for the Cortex host server.
// Values are loaded from environment variables with sensible defaults for local development.
type Config struct {
	Port        int
	DataDir     string
	PluginDir   string
	FrontendDir string
}

// Load reads configuration from environment variables and validates it.
// It returns an error immediately if any required value is invalid,
// following the fail-fast principle.
func Load() (*Config, error) {
	config := &Config{
		Port:        getEnvAsInt("CORTEX_PORT", 8080),
		DataDir:     getEnv("CORTEX_DATA_DIR", "./data"),
		PluginDir:   getEnv("CORTEX_PLUGIN_DIR", "./plugins"),
		FrontendDir: getEnv("CORTEX_FRONTEND_DIR", "./frontend/build"),
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// validate checks that all configuration values are within acceptable bounds.
func (c *Config) validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("CORTEX_PORT must be between 1 and 65535, got %d", c.Port)
	}

	if c.DataDir == "" {
		return fmt.Errorf("CORTEX_DATA_DIR must not be empty")
	}

	if c.PluginDir == "" {
		return fmt.Errorf("CORTEX_PLUGIN_DIR must not be empty")
	}

	if c.FrontendDir == "" {
		return fmt.Errorf("CORTEX_FRONTEND_DIR must not be empty")
	}

	return nil
}

// Address returns the formatted listen address for the HTTP server.
func (c *Config) Address() string {
	return fmt.Sprintf(":%d", c.Port)
}

// getEnv reads an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt reads an environment variable as an integer or returns a default value.
// If the value cannot be parsed as an integer, the default is returned.
func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}
