package fs

import (
	"io/fs"
	"os"
	"path/filepath"
)

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
