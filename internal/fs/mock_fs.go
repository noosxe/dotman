package fs

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing/fstest"
)

// MockFileSystem implements FileSystem for testing
type MockFileSystem struct {
	rootDir string
	homeDir string
}

// NewMockFileSystem creates a new MockFileSystem
func NewMockFileSystem(files map[string]*fstest.MapFile) (*MockFileSystem, error) {
	return NewMockFileSystemWithHome(files, "/home/test")
}

// NewMockFileSystemWithHome creates a new MockFileSystem with a custom home directory
func NewMockFileSystemWithHome(files map[string]*fstest.MapFile, homeDir string) (*MockFileSystem, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}

	if files == nil {
		files = make(map[string]*fstest.MapFile)
	}

	fs := &MockFileSystem{
		rootDir: tmpDir,
		homeDir: homeDir,
	}

	for key, val := range files {
		dir := filepath.Dir(key)
		err = fs.MkdirAll(dir, 0755)
		if err != nil {
			return nil, err
		}

		err = fs.WriteFile(key, val.Data, val.Mode)
		if err != nil {
			return nil, err
		}
	}

	return fs, nil
}

// Cleanup the temp directory
func (m *MockFileSystem) CleanUp() {
	os.RemoveAll(m.rootDir)
}

func (m *MockFileSystem) DumpTree() string {
	builder := strings.Builder{}
	filepath.WalkDir(m.rootDir, func(name string, d os.DirEntry, err error) error {
		builder.WriteString(name)
		builder.WriteString("\n")
		return nil
	})

	return builder.String()
}

// MkdirAll creates directories in the mock filesystem
func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	dirPath := filepath.Join(m.rootDir, path)
	return os.MkdirAll(dirPath, perm)
}

// WriteFile adds a file to the mock filesystem
func (m *MockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	filePath := filepath.Join(m.rootDir, name)
	return os.WriteFile(filePath, data, perm)
}

// ReadFile reads a file from the mock filesystem
func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	filePath := filepath.Join(m.rootDir, name)
	return os.ReadFile(filePath)
}

// Remove removes a file from the mock filesystem
func (m *MockFileSystem) Remove(name string) error {
	filePath := filepath.Join(m.rootDir, name)
	return os.Remove(filePath)
}

// RemoveAll implements FileSystem
func (m *MockFileSystem) RemoveAll(path string) error {
	fullPath := filepath.Join(m.rootDir, path)
	return os.RemoveAll(fullPath)
}

// Symlink implements FileSystem
func (m *MockFileSystem) Symlink(oldname, newname string) error {
	old := filepath.Join(m.rootDir, oldname)
	new := filepath.Join(m.rootDir, newname)
	return os.Symlink(old, new)
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
	// Clean the paths to handle any ".." or "." components
	basepath = filepath.Clean(basepath)
	targpath = filepath.Clean(targpath)

	// Check if target path is under base path
	rel, err := filepath.Rel(basepath, targpath)
	if err != nil {
		return "", err
	}
	if rel == ".." || len(rel) > 2 && rel[:3] == ".."+string(filepath.Separator) {
		return "", os.ErrInvalid
	}
	return rel, nil
}

// Stat implements fs.StatFS
func (m *MockFileSystem) Stat(name string) (fs.FileInfo, error) {
	filePath := filepath.Join(m.rootDir, name)
	return os.Stat(filePath)
}

// Open implements fs.FS
func (m *MockFileSystem) Open(name string) (*os.File, error) {
	filePath := filepath.Join(m.rootDir, name)
	return os.Open(filePath)
}

func (m *MockFileSystem) RealPath(path string) string {
	return filepath.Join(m.rootDir, path)
}

func (m *MockFileSystem) Readdir(path string) ([]os.FileInfo, error) {
	dir, err := m.Open(path)
	if err != nil {
		return nil, err
	}

	return dir.Readdir(0)
}
