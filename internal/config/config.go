package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	dotmanfs "github.com/noosxe/dotman/internal/fs"
)

// Config represents the dotman configuration
type Config struct {
	DotmanDir string `json:"dotman_dir"`
}

// DefaultConfig returns the default configuration
func DefaultConfig(fsys dotmanfs.FileSystem) *Config {
	home, err := fsys.UserHomeDir()
	if err != nil {
		home = "~"
	}
	return &Config{
		DotmanDir: filepath.Join(home, ".dotman"),
	}
}

// LoadConfig loads the configuration from the specified path
func LoadConfig(configPath string, fsys dotmanfs.FileSystem) (*Config, error) {
	fmt.Printf("Loading config from: %s\n", configPath)

	// Check if config file exists
	if _, err := fsys.Stat(configPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error checking config file: %v", err)
		}
		fmt.Printf("Config file does not exist, creating default config\n")
		// Create default config if it doesn't exist
		config := DefaultConfig(fsys)
		if err := SaveConfig(configPath, config, fsys); err != nil {
			return nil, fmt.Errorf("error creating default config: %v", err)
		}
		return config, nil
	}

	// Read and parse config file
	data, err := fsys.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the specified path
func SaveConfig(configPath string, config *Config, fsys dotmanfs.FileSystem) error {
	fmt.Printf("Saving config to: %s\n", configPath)

	// Ensure the directory exists
	dir := filepath.Dir(configPath)
	if err := fsys.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	if err := fsys.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}
