package fs

import (
	"io/fs"
	"os"
	"testing/fstest"
	"time"
)

// FileSystem is an interface that combines read and write filesystem operations
type FileSystem interface {
	// Read operations
	fs.FS
	fs.StatFS
	ReadFile(name string) ([]byte, error)

	// Write operations
	MkdirAll(path string, perm os.FileMode) error
	WriteFile(name string, data []byte, perm os.FileMode) error
	Remove(name string) error

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

// ReadFile implements FileSystem
func (f *OSFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// MkdirAll implements FileSystem
func (f *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// WriteFile implements FileSystem
func (f *OSFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// Remove implements FileSystem
func (f *OSFileSystem) Remove(name string) error {
	return os.Remove(name)
}

// UserHomeDir implements FileSystem
func (f *OSFileSystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// MockFileSystem implements FileSystem for testing
type MockFileSystem struct {
	fstest.MapFS
	homeDir string
}

// NewMockFileSystem creates a new MockFileSystem
func NewMockFileSystem(files map[string]*fstest.MapFile) *MockFileSystem {
	if files == nil {
		files = make(map[string]*fstest.MapFile)
	}
	mfs := fstest.MapFS(files)
	return &MockFileSystem{
		MapFS:   mfs,
		homeDir: "/home/test",
	}
}

// MkdirAll creates directories in the mock filesystem
func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	m.MapFS[path] = &fstest.MapFile{
		Data: []byte{},
		Mode: perm | fs.ModeDir,
	}
	return nil
}

// WriteFile adds a file to the mock filesystem
func (m *MockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	m.MapFS[name] = &fstest.MapFile{
		Data: data,
		Mode: perm,
	}
	return nil
}

// ReadFile reads a file from the mock filesystem
func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	if file, ok := m.MapFS[name]; ok {
		return file.Data, nil
	}
	return nil, os.ErrNotExist
}

// Remove removes a file from the mock filesystem
func (m *MockFileSystem) Remove(name string) error {
	delete(m.MapFS, name)
	return nil
}

// UserHomeDir implements FileSystem
func (m *MockFileSystem) UserHomeDir() (string, error) {
	return m.homeDir, nil
}

// mapFileInfo wraps fstest.MapFile to implement fs.FileInfo
type mapFileInfo struct {
	*fstest.MapFile
	name string
}

func (m *mapFileInfo) Name() string       { return m.name }
func (m *mapFileInfo) Size() int64        { return int64(len(m.Data)) }
func (m *mapFileInfo) Mode() fs.FileMode  { return m.MapFile.Mode }
func (m *mapFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mapFileInfo) IsDir() bool        { return m.Mode()&fs.ModeDir != 0 }
func (m *mapFileInfo) Sys() interface{}   { return nil }

// Stat implements fs.StatFS
func (m *MockFileSystem) Stat(name string) (fs.FileInfo, error) {
	if file, ok := m.MapFS[name]; ok {
		return &mapFileInfo{
			MapFile: file,
			name:    name,
		}, nil
	}
	return nil, os.ErrNotExist
}
