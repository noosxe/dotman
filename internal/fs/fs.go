package fs

import (
	"io/fs"
	"os"
	"path/filepath"
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
	RemoveAll(path string) error
	Symlink(oldname, newname string) error

	// User operations
	UserHomeDir() (string, error)

	// Path operations
	Abs(path string) (string, error)
	Rel(basepath, targpath string) (string, error)
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

// RemoveAll implements FileSystem
func (f *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Symlink implements FileSystem
func (f *OSFileSystem) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

// UserHomeDir implements FileSystem
func (f *OSFileSystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// Abs implements FileSystem
func (f *OSFileSystem) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// Rel implements FileSystem
func (f *OSFileSystem) Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath)
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
