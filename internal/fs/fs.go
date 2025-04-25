package fs

import (
	"io/fs"
	"os"
	"testing/fstest"
)

// FileSystem is an interface that combines read and write filesystem operations
type FileSystem interface {
	// Read operations
	fs.FS
	fs.StatFS

	// Write operations
	MkdirAll(path string, perm os.FileMode) error
	WriteFile(name string, data []byte, perm os.FileMode) error

	// User operations
	UserHomeDir() (string, error)
}

// OSFileSystem implements FileSystem using the real filesystem
type OSFileSystem struct{}

// NewOSFileSystem creates a new OSFileSystem instance
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// Open implements fs.FS
func (f *OSFileSystem) Open(name string) (fs.File, error) {
	return os.Open(name)
}

// Stat implements fs.StatFS
func (f *OSFileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

// MkdirAll implements FileSystem
func (f *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// WriteFile implements FileSystem
func (f *OSFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// UserHomeDir implements FileSystem
func (f *OSFileSystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// MockFileSystem implements FileSystem for testing
type MockFileSystem struct {
	*fstest.MapFS
	homeDir string
}

// NewMockFileSystem creates a new MockFileSystem
func NewMockFileSystem(files map[string]*fstest.MapFile) *MockFileSystem {
	if files == nil {
		files = make(map[string]*fstest.MapFile)
	}
	mfs := fstest.MapFS(files)
	return &MockFileSystem{
		MapFS:   &mfs,
		homeDir: "/home/test",
	}
}

// MkdirAll is a no-op for testing
func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

// WriteFile adds a file to the mock filesystem
func (m *MockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	(*m.MapFS)[name] = &fstest.MapFile{
		Data: data,
		Mode: perm,
	}
	return nil
}

// UserHomeDir implements FileSystem
func (m *MockFileSystem) UserHomeDir() (string, error) {
	return m.homeDir, nil
}
