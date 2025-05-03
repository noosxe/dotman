package cmd

import (
	"context"
	"path/filepath"
	"testing"
	stdFstest "testing/fstest"

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

	// Get entry from context to verify initialization
	entry, err := journal.GetJournalEntry(op.ctx)
	if err != nil {
		t.Errorf("failed to get journal entry: %v", err)
	}

	if entry.Operation != journal.OperationTypeAdd {
		t.Errorf("expected operation '%s', got '%s'", journal.OperationTypeAdd, entry.Operation)
	}

	if entry.Source != "test/file" {
		t.Errorf("expected source 'test/file', got '%s'", entry.Source)
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

			// Initialize operation
			op := &addOperation{
				path: tt.path,
				fsys: mockFS,
				ctx:  context.Background(),
			}

			// Set up journal manager and entry in context
			jm := journal.NewJournalManager(mockFS, "dotman/journal")
			entry, err := jm.CreateEntry(journal.OperationTypeAdd, tt.path, "")
			if err != nil {
				t.Fatalf("failed to create journal entry: %v", err)
			}

			op.ctx = journal.WithJournalManager(op.ctx, jm)
			op.ctx = journal.WithJournalEntry(op.ctx, entry)

			err = op.verifySource()
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				// Get updated entry to check state
				entry, err := journal.GetJournalEntry(op.ctx)
				if err != nil {
					t.Fatalf("failed to get journal entry: %v", err)
				}
				if entry.State != journal.EntryStateFailed {
					t.Error("expected entry to be moved to failed state")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Get updated entry to check steps
				entry, err := journal.GetJournalEntry(op.ctx)
				if err != nil {
					t.Fatalf("failed to get journal entry: %v", err)
				}
				if len(entry.Steps) != 1 {
					t.Errorf("expected 1 step, got %d", len(entry.Steps))
				}
				step := entry.Steps[0]
				if step.Type != journal.StepTypeVerify || step.Status != journal.StepStatusCompleted {
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

	// Initialize operation
	op := &addOperation{
		path: sourcePath,
		fsys: mockFS,
		ctx:  context.Background(),
	}

	// Set up journal manager and entry in context
	jm := journal.NewJournalManager(mockFS, "dotman/journal")
	entry, err := jm.CreateEntry(journal.OperationTypeAdd, sourcePath, targetPath)
	if err != nil {
		t.Fatalf("failed to create journal entry: %v", err)
	}

	op.ctx = journal.WithJournalManager(op.ctx, jm)
	op.ctx = journal.WithJournalEntry(op.ctx, entry)

	err = op.copyAndVerifyFile(targetPath)
	if err != nil {
		t.Errorf("copyAndVerifyFile() returned error: %v", err)
	}

	// Verify file was copied
	if _, err := mockFS.Stat(targetPath); err != nil {
		t.Errorf("target file was not created: %v", err)
	}

	// Get updated entry to verify journal steps
	entry, err = journal.GetJournalEntry(op.ctx)
	if err != nil {
		t.Fatalf("failed to get journal entry: %v", err)
	}

	if len(entry.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(entry.Steps))
	}

	// Check copy step
	copyStep := entry.Steps[0]
	if copyStep.Type != journal.StepTypeCopy || copyStep.Status != journal.StepStatusCompleted {
		t.Errorf("unexpected copy step: %+v", copyStep)
	}

	// Check verify step
	verifyStep := entry.Steps[1]
	if verifyStep.Type != journal.StepTypeVerify || verifyStep.Status != journal.StepStatusCompleted {
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

	// Initialize operation
	op := &addOperation{
		path: sourcePath,
		fsys: mockFS,
		ctx:  context.Background(),
	}

	// Set up journal manager and entry in context
	jm := journal.NewJournalManager(mockFS, "dotman/journal")
	entry, err := jm.CreateEntry(journal.OperationTypeAdd, sourcePath, targetPath)
	if err != nil {
		t.Fatalf("failed to create journal entry: %v", err)
	}

	op.ctx = journal.WithJournalManager(op.ctx, jm)
	op.ctx = journal.WithJournalEntry(op.ctx, entry)

	err = op.copyAndVerifyDirectory(targetPath)
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

	// Get updated entry to verify journal steps
	entry, err = journal.GetJournalEntry(op.ctx)
	if err != nil {
		t.Fatalf("failed to get journal entry: %v", err)
	}

	if len(entry.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(entry.Steps))
	}

	// Check copy step
	copyStep := entry.Steps[0]
	if copyStep.Type != journal.StepTypeCopy || copyStep.Status != journal.StepStatusCompleted {
		t.Errorf("unexpected copy step: %+v", copyStep)
	}

	// Check verify step
	verifyStep := entry.Steps[1]
	if verifyStep.Type != journal.StepTypeVerify || verifyStep.Status != journal.StepStatusCompleted {
		t.Errorf("unexpected verify step: %+v", verifyStep)
	}
}

func TestAddOperation_Complete(t *testing.T) {
	mockFS := dotmanfs.NewMockFileSystem(nil)

	// Initialize operation
	op := &addOperation{
		fsys: mockFS,
		ctx:  context.Background(),
	}

	// Set up journal manager and entry in context
	jm := journal.NewJournalManager(mockFS, "dotman/journal")
	entry, err := jm.CreateEntry(journal.OperationTypeAdd, "", "")
	if err != nil {
		t.Fatalf("failed to create journal entry: %v", err)
	}

	op.ctx = journal.WithJournalManager(op.ctx, jm)
	op.ctx = journal.WithJournalEntry(op.ctx, entry)

	err = op.complete()
	if err != nil {
		t.Errorf("complete() returned error: %v", err)
	}

	// Get updated entry to check state
	entry, err = journal.GetJournalEntry(op.ctx)
	if err != nil {
		t.Fatalf("failed to get journal entry: %v", err)
	}

	if entry.State != journal.EntryStateCompleted {
		t.Errorf("expected state '%s', got '%s'", journal.EntryStateCompleted, entry.State)
	}
}

func TestAddOperation_CreateSymlink(t *testing.T) {
	sourcePath := "test/source"
	targetPath := "dotman/source"

	// Create mock file system
	mockFS := dotmanfs.NewMockFileSystem(nil)

	// Add both source and target files
	if err := mockFS.WriteFile(sourcePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}
	if err := mockFS.WriteFile(targetPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create target file: %v", err)
	}

	// Initialize operation
	op := &addOperation{
		path: sourcePath,
		fsys: mockFS,
		ctx:  context.Background(),
		config: &config.Config{
			DotmanDir: "dotman",
		},
	}

	// Set up journal manager and entry in context
	jm := journal.NewJournalManager(mockFS, "dotman/journal")
	entry, err := jm.CreateEntry(journal.OperationTypeAdd, sourcePath, targetPath)
	if err != nil {
		t.Fatalf("failed to create journal entry: %v", err)
	}

	op.ctx = journal.WithJournalManager(op.ctx, jm)
	op.ctx = journal.WithJournalEntry(op.ctx, entry)

	err = op.createSymlink()
	if err != nil {
		t.Errorf("createSymlink() returned error: %v", err)
	}

	// Verify source path is now a symlink
	if _, err := mockFS.Stat(sourcePath); err != nil {
		t.Errorf("symlink was not created: %v", err)
	}

	// Verify target path still exists
	if _, err := mockFS.Stat(targetPath); err != nil {
		t.Errorf("target file was removed: %v", err)
	}

	// Get updated entry to verify journal steps
	entry, err = journal.GetJournalEntry(op.ctx)
	if err != nil {
		t.Fatalf("failed to get journal entry: %v", err)
	}

	if len(entry.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(entry.Steps))
	}

	// Check symlink step
	symlinkStep := entry.Steps[0]
	if symlinkStep.Type != journal.StepTypeSymlink || symlinkStep.Status != journal.StepStatusCompleted {
		t.Errorf("unexpected symlink step: %+v", symlinkStep)
	}
}
