package testutil

import (
	"path/filepath"

	dotmanfs "github.com/noosxe/dotman/internal/fs"
)

const (
	// TestHomeDir is the default home directory used in tests
	TestHomeDir = "/home/test"
)

// NewMockFS creates a new mock filesystem with a home directory at /home/test
func NewMockFS() dotmanfs.FileSystem {
	fsys := dotmanfs.NewMockFileSystemWithHome(nil, TestHomeDir)
	// Create home directory
	fsys.MkdirAll(TestHomeDir, 0755)
	return fsys
}

// NewMockFSWithDotman creates a new mock filesystem with a home directory and dotman directory structure
func NewMockFSWithDotman() (dotmanfs.FileSystem, string) {
	fsys := NewMockFS()

	// Create dotman directory
	dotmanDir := filepath.Join(TestHomeDir, ".dotman")
	fsys.MkdirAll(dotmanDir, 0755)

	// Create data directory
	fsys.MkdirAll(filepath.Join(dotmanDir, "data"), 0755)

	// Create journal directory and its subdirectories
	journalDir := filepath.Join(dotmanDir, "journal")
	fsys.MkdirAll(journalDir, 0755)
	for _, subdir := range []string{"current", "completed", "failed"} {
		fsys.MkdirAll(filepath.Join(journalDir, subdir), 0755)
	}

	return fsys, dotmanDir
}
