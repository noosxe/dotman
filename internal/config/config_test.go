package config

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/noosxe/dotman/internal/fs"
)

func TestLoadConfig_NewConfig(t *testing.T) {
	// Create a mock filesystem
	mockFS, err := fs.NewMockFileSystem(nil)
	if err != nil {
		t.Fatalf("failed to create mock filesystem: %v", err)
	}
	defer mockFS.CleanUp()

	configPath := "config.json"

	cfg, err := LoadConfig(configPath, mockFS)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	expectedDir := filepath.Join("/home/test", ".dotman")
	if cfg.DotmanDir != expectedDir {
		t.Errorf("Expected DotmanDir to be %s, got %s", expectedDir, cfg.DotmanDir)
	}
}

func TestLoadConfig_ExistingConfig(t *testing.T) {
	// Create a filesystem with existing config
	existingConfig := &Config{
		DotmanDir: "/custom/dotman/dir",
	}
	data, err := json.Marshal(existingConfig)
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	mockFS, err := fs.NewMockFileSystem(map[string]*fstest.MapFile{
		"config.json": {
			Data: data,
		},
	})
	if err != nil {
		t.Fatalf("failed to create mock filesystem: %v", err)
	}
	defer mockFS.CleanUp()

	configPath := "config.json"
	cfg, err := LoadConfig(configPath, mockFS)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.DotmanDir != existingConfig.DotmanDir {
		t.Errorf("Expected DotmanDir to be %s, got %s", existingConfig.DotmanDir, cfg.DotmanDir)
	}
}

func TestSaveConfig(t *testing.T) {
	mockFS, err := fs.NewMockFileSystem(nil)
	if err != nil {
		t.Fatalf("failed to create mock filesystem: %v", err)
	}
	defer mockFS.CleanUp()

	configPath := "config.json"
	cfg := &Config{
		DotmanDir: "/test/dotman",
	}

	err = SaveConfig(configPath, cfg, mockFS)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify the saved data
	data, err := mockFS.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	var savedConfig Config
	if err := json.Unmarshal(data, &savedConfig); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	if savedConfig.DotmanDir != cfg.DotmanDir {
		t.Errorf("Expected saved DotmanDir to be %s, got %s", cfg.DotmanDir, savedConfig.DotmanDir)
	}
}
