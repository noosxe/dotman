package fs

import (
	"os"
)

// FileSystem is an interface that combines read and write filesystem operations
type FileSystem interface {
	// Read operations
	Open(file string) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
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
	Readdir(path string) ([]os.FileInfo, error)
}
