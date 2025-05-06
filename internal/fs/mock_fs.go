package fs

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// Open implements fs.FS
func (m *MockFileSystem) Open(name string) (fs.File, error) {
	if file, ok := m.MapFS[name]; ok {
		return &mockFile{
			name:   name,
			data:   file.Data,
			mode:   file.Mode,
			offset: 0,
			fs:     m,
		}, nil
	}
	return nil, os.ErrNotExist
}

// mockFile implements fs.File
type mockFile struct {
	name   string
	data   []byte
	mode   fs.FileMode
	offset int64
	fs     *MockFileSystem
}

// Stat implements fs.File
func (f *mockFile) Stat() (fs.FileInfo, error) {
	return &mapFileInfo{
		MapFile: &fstest.MapFile{
			Data: f.data,
			Mode: f.mode,
		},
		name: f.name,
	}, nil
}

// Read implements fs.File
func (f *mockFile) Read(p []byte) (n int, err error) {
	if f.offset >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n = copy(p, f.data[f.offset:])
	f.offset += int64(n)
	return n, nil
}

// Close implements fs.File
func (f *mockFile) Close() error {
	return nil
}

// mockDirEntry implements fs.DirEntry
type mockDirEntry struct {
	name string
	mode fs.FileMode
}

func (d *mockDirEntry) Name() string               { return d.name }
func (d *mockDirEntry) IsDir() bool                { return d.mode.IsDir() }
func (d *mockDirEntry) Type() fs.FileMode          { return d.mode.Type() }
func (d *mockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// ReadDir implements fs.ReadDirFile
func (f *mockFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if !f.mode.IsDir() {
		return nil, os.ErrInvalid
	}

	// Find all files under this directory
	var entries []fs.DirEntry
	dirPath := f.name
	if !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}

	for name, file := range f.fs.MapFS {
		// Skip if not under this directory
		if !strings.HasPrefix(name, dirPath) {
			continue
		}

		// Get the relative path
		relPath := strings.TrimPrefix(name, dirPath)
		if relPath == "" {
			continue
		}

		// Only include immediate children
		if strings.Contains(relPath, "/") {
			continue
		}

		entries = append(entries, &mockDirEntry{
			name: relPath,
			mode: file.Mode,
		})
	}

	// Sort entries by name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	// Handle n parameter
	if n > 0 && len(entries) > n {
		entries = entries[:n]
	}

	return entries, nil
}
