package fs

import (
	"io/fs"
	"os"
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
