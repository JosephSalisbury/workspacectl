package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// WindowConfig defines a window in a tmux layout.
type WindowConfig struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

// Config holds the workspacectl configuration.
type Config struct {
	Layouts map[string][]WindowConfig `yaml:"layouts"`
}

var defaultConfig = Config{
	Layouts: map[string][]WindowConfig{
		"worktree": {
			{Name: "claude", Command: "claude"},
			{Name: "diff", Command: "watch -n 5 git diff"},
		},
		"temporary": {
			{Name: "claude", Command: "claude"},
		},
	},
}

// LoadConfig reads a config file from disk.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// EnsureConfig creates the base dir and default config if they don't exist.
func EnsureConfig(baseDir string) (Config, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return Config{}, fmt.Errorf("creating base dir: %w", err)
	}

	configPath := filepath.Join(baseDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		data, err := yaml.Marshal(defaultConfig)
		if err != nil {
			return Config{}, fmt.Errorf("marshalling default config: %w", err)
		}
		if err := os.WriteFile(configPath, data, 0o644); err != nil {
			return Config{}, fmt.Errorf("writing default config: %w", err)
		}
		return defaultConfig, nil
	}

	return LoadConfig(configPath)
}
