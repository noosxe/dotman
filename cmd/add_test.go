package cmd

import (
	"path/filepath"
	"testing"
	stdFstest "testing/fstest"
	"time"

	"github.com/noosxe/dotman/internal/config"
	dotmanfs "github.com/noosxe/dotman/internal/fs"
	"github.com/noosxe/dotman/internal/journal"
)

func TestAddOperation_Initialize(t *testing.T) {
	mockFS := dotmanfs.NewMockFileSystem(nil)
	op := &addOperation{
		path: "test/file",
		fsys: mockFS,
	}

	err := op.initialize()
	if err != nil {
		t.Errorf("initialize() returned error: %v", err)
	}

	if op.config == nil {
		t.Error("config was not initialized")
	}

	if op.journalManager == nil {
		t.Error("journalManager was not initialized")
	}

	if op.entry == nil {
		t.Error("journal entry was not created")
	}

	if op.entry.Operation != "add" {
		t.Errorf("expected operation 'add', got '%s'", op.entry.Operation)
	}

	if op.entry.Source != "test/file" {
		t.Errorf("expected source 'test/file', got '%s'", op.entry.Source)
	}
}

func TestAddOperation_VerifySource(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		exists      bool
		expectError bool
	}{
		{
			name:        "file exists",
			path:        "test/file",
			exists:      true,
			expectError: false,
		},
		{
			name:        "file does not exist",
			path:        "test/nonexistent",
			exists:      false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock file system with initial state
			initialState := make(map[string]*stdFstest.MapFile)
			if tt.exists {
				initialState[tt.path] = &stdFstest.MapFile{
					Data: []byte{},
					Mode: 0644,
				}
			}
			mockFS := dotmanfs.NewMockFileSystem(initialState)

			op := &addOperation{
				path:           tt.path,
				fsys:           mockFS,
				config:         &config.Config{DotmanDir: "dotman"},
				journalManager: journal.NewJournalManager(mockFS, "dotman/journal"),
				entry: &journal.JournalEntry{
					ID:        "test",
					Timestamp: time.Now(),
					Operation: "add",
					State:     "current",
					Steps:     make([]journal.Step, 0),
				},
			}

			err := op.verifySource()
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if op.entry.State != "failed" {
					t.Error("expected entry to be moved to failed state")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(op.entry.Steps) != 1 {
					t.Errorf("expected 1 step, got %d", len(op.entry.Steps))
				}
				step := op.entry.Steps[0]
				if step.Type != "verify" || step.Status != "completed" {
					t.Errorf("unexpected step: %+v", step)
				}
			}
		})
	}
}

func TestAddOperation_CopyAndVerifyFile(t *testing.T) {
	sourcePath := "test/source"
	targetPath := "dotman/source"

	// Create mock file system
	mockFS := dotmanfs.NewMockFileSystem(nil)

	// Add source file
	if err := mockFS.WriteFile(sourcePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	op := &addOperation{
		path:           sourcePath,
		fsys:           mockFS,
		config:         &config.Config{DotmanDir: "dotman"},
		journalManager: journal.NewJournalManager(mockFS, "dotman/journal"),
		entry: &journal.JournalEntry{
			ID:        "test",
			Timestamp: time.Now(),
			Operation: "add",
			State:     "current",
			Steps:     make([]journal.Step, 0),
		},
	}

	err := op.copyAndVerifyFile(targetPath)
	if err != nil {
		t.Errorf("copyAndVerifyFile() returned error: %v", err)
	}

	// Verify file was copied
	if _, err := mockFS.Stat(targetPath); err != nil {
		t.Errorf("target file was not created: %v", err)
	}

	// Verify journal steps
	if len(op.entry.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(op.entry.Steps))
	}

	// Check copy step
	copyStep := op.entry.Steps[0]
	if copyStep.Type != "copy" || copyStep.Status != "completed" {
		t.Errorf("unexpected copy step: %+v", copyStep)
	}

	// Check verify step
	verifyStep := op.entry.Steps[1]
	if verifyStep.Type != "verify" || verifyStep.Status != "completed" {
		t.Errorf("unexpected verify step: %+v", verifyStep)
	}
}

func TestAddOperation_CopyAndVerifyDirectory(t *testing.T) {
	mockFS := dotmanfs.NewMockFileSystem(nil)
	sourcePath := "test/source"
	targetPath := "dotman/source"

	// Create source directory structure
	mockFS.MkdirAll(sourcePath, 0755)
	mockFS.WriteFile(filepath.Join(sourcePath, "file1"), []byte("test1"), 0644)
	mockFS.WriteFile(filepath.Join(sourcePath, "file2"), []byte("test2"), 0644)
	mockFS.MkdirAll(filepath.Join(sourcePath, "subdir"), 0755)
	mockFS.WriteFile(filepath.Join(sourcePath, "subdir", "file3"), []byte("test3"), 0644)

	op := &addOperation{
		path:           sourcePath,
		fsys:           mockFS,
		config:         &config.Config{DotmanDir: "dotman"},
		journalManager: journal.NewJournalManager(mockFS, "dotman/journal"),
		entry: &journal.JournalEntry{
			ID:        "test",
			Timestamp: time.Now(),
			Operation: "add",
			State:     "current",
			Steps:     make([]journal.Step, 0),
		},
	}

	err := op.copyAndVerifyDirectory(targetPath)
	if err != nil {
		t.Errorf("copyAndVerifyDirectory() returned error: %v", err)
	}

	// Verify directory structure was copied
	verifyPaths := []string{
		targetPath,
		filepath.Join(targetPath, "file1"),
		filepath.Join(targetPath, "file2"),
		filepath.Join(targetPath, "subdir"),
		filepath.Join(targetPath, "subdir", "file3"),
	}

	for _, path := range verifyPaths {
		if _, err := mockFS.Stat(path); err != nil {
			t.Errorf("path %s was not created: %v", path, err)
		}
	}

	// Verify journal steps
	if len(op.entry.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(op.entry.Steps))
	}

	// Check copy step
	copyStep := op.entry.Steps[0]
	if copyStep.Type != "copy" || copyStep.Status != "completed" {
		t.Errorf("unexpected copy step: %+v", copyStep)
	}

	// Check verify step
	verifyStep := op.entry.Steps[1]
	if verifyStep.Type != "verify" || verifyStep.Status != "completed" {
		t.Errorf("unexpected verify step: %+v", verifyStep)
	}
}

func TestAddOperation_Complete(t *testing.T) {
	mockFS := dotmanfs.NewMockFileSystem(nil)
	op := &addOperation{
		fsys:           mockFS,
		config:         &config.Config{DotmanDir: "dotman"},
		journalManager: journal.NewJournalManager(mockFS, "dotman/journal"),
		entry: &journal.JournalEntry{
			ID:        "test",
			Timestamp: time.Now(),
			Operation: "add",
			State:     "current",
			Steps:     make([]journal.Step, 0),
		},
	}

	err := op.complete()
	if err != nil {
		t.Errorf("complete() returned error: %v", err)
	}

	if op.entry.State != "completed" {
		t.Errorf("expected state 'completed', got '%s'", op.entry.State)
	}
}
