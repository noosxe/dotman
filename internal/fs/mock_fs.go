package fs

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing/fstest"
	"time"
)

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

// NewMockFileSystemWithHome creates a new MockFileSystem with a custom home directory
func NewMockFileSystemWithHome(files map[string]*fstest.MapFile, homeDir string) *MockFileSystem {
	if files == nil {
		files = make(map[string]*fstest.MapFile)
	}
	mfs := fstest.MapFS(files)
	return &MockFileSystem{
		MapFS:   mfs,
		homeDir: homeDir,
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

// RemoveAll implements FileSystem
func (m *MockFileSystem) RemoveAll(path string) error {
	// Remove all files that start with the path
	for name := range m.MapFS {
		if name == path || filepath.Dir(name) == path {
			delete(m.MapFS, name)
		}
	}
	return nil
}

// Symlink implements FileSystem
func (m *MockFileSystem) Symlink(oldname, newname string) error {
	// In the mock filesystem, we'll just copy the file
	if file, ok := m.MapFS[oldname]; ok {
		m.MapFS[newname] = file
		return nil
	}
	return os.ErrNotExist
}

// UserHomeDir implements FileSystem
func (m *MockFileSystem) UserHomeDir() (string, error) {
	return m.homeDir, nil
}

// Abs implements FileSystem
func (m *MockFileSystem) Abs(path string) (string, error) {
	// In the mock filesystem, we'll just return the path as is
	return path, nil
}

// Rel implements FileSystem
func (m *MockFileSystem) Rel(basepath, targpath string) (string, error) {
	// In the mock filesystem, we'll just return the path as is
	return targpath, nil
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
