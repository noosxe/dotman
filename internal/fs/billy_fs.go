package fs

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-billy/v5"
)

// BillyFileSystem adapts our FileSystem interface to go-billy's filesystem interface
type BillyFileSystem struct {
	fs       FileSystem
	basePath string
}

// NewBillyFileSystem creates a new BillyFileSystem instance
func NewBillyFileSystem(fs FileSystem, basePath string) *BillyFileSystem {
	return &BillyFileSystem{
		fs:       fs,
		basePath: filepath.Join(basePath, ""),
	}
}

// Create implements billy.Filesystem
func (b *BillyFileSystem) Create(filename string) (billy.File, error) {
	return b.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Open implements billy.Filesystem
func (b *BillyFileSystem) Open(filename string) (billy.File, error) {
	return b.OpenFile(filename, os.O_RDONLY, 0)
}

// OpenFile implements billy.Filesystem
func (b *BillyFileSystem) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	filePath := filepath.Join(b.basePath, filename)
	// Create parent directories if needed
	if flag&os.O_CREATE != 0 {
		dir := filepath.Dir(filePath)
		if err := b.fs.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// Read existing file if it exists
	var data []byte
	var err error
	if flag&os.O_CREATE == 0 {
		data, err = b.fs.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, err
			}
			return nil, err
		}
	}

	return &billyFile{
		fs:       b.fs,
		name:     filename,
		data:     data,
		flag:     flag,
		perm:     perm,
		offset:   0,
		basePath: b.basePath,
	}, nil
}

// Stat implements billy.Filesystem
func (b *BillyFileSystem) Stat(filename string) (os.FileInfo, error) {
	return b.fs.Stat(filepath.Join(b.basePath, filename))
}

// Rename implements billy.Filesystem
func (b *BillyFileSystem) Rename(oldpath, newpath string) error {
	// Read the old file
	old := filepath.Join(b.basePath, oldpath)
	data, err := b.fs.ReadFile(old)
	if err != nil {
		return err
	}

	// Write to the new file
	new := filepath.Join(b.basePath, newpath)
	if err := b.fs.MkdirAll(filepath.Dir(new), 0755); err != nil {
		return err
	}
	if err := b.fs.WriteFile(new, data, 0644); err != nil {
		return err
	}

	// Remove the old file
	return b.fs.Remove(old)
}

// Remove implements billy.Filesystem
func (b *BillyFileSystem) Remove(filename string) error {
	return b.fs.Remove(filepath.Join(b.basePath, filename))
}

// Join implements billy.Filesystem
func (b *BillyFileSystem) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// TempFile implements billy.Filesystem
func (b *BillyFileSystem) TempFile(dir, prefix string) (billy.File, error) {
	// Create a temporary file name
	name := filepath.Join(dir, prefix+time.Now().Format("20060102150405"))
	return b.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
}

// ReadDir implements billy.Filesystem
func (b *BillyFileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	return b.fs.Readdir(filepath.Join(b.basePath, path))
}

// MkdirAll implements billy.Filesystem
func (b *BillyFileSystem) MkdirAll(filename string, perm os.FileMode) error {
	return b.fs.MkdirAll(filepath.Join(b.basePath, filename), perm)
}

// Lstat implements billy.Filesystem
func (b *BillyFileSystem) Lstat(filename string) (os.FileInfo, error) {
	return b.fs.Stat(filepath.Join(b.basePath, filename))
}

// Symlink implements billy.Filesystem
func (b *BillyFileSystem) Symlink(target, link string) error {
	return b.fs.Symlink(filepath.Join(b.basePath, target), filepath.Join(b.basePath, link))
}

// Readlink implements billy.Filesystem
func (b *BillyFileSystem) Readlink(link string) (string, error) {
	// For now, we'll just return the link as is
	// TODO: Implement proper symlink resolution
	return link, nil
}

// Chroot implements billy.Filesystem
func (b *BillyFileSystem) Chroot(path string) (billy.Filesystem, error) {
	return &BillyFileSystem{
		fs:       b.fs,
		basePath: filepath.Join(b.basePath, path),
	}, nil
}

// Root implements billy.Filesystem
func (b *BillyFileSystem) Root() string {
	return b.basePath
}

// Capabilities implements billy.Filesystem
func (b *BillyFileSystem) Capabilities() billy.Capability {
	return billy.ReadCapability | billy.WriteCapability | billy.ReadAndWriteCapability
}

// billyFile implements billy.File
type billyFile struct {
	fs       FileSystem
	name     string
	data     []byte
	flag     int
	perm     os.FileMode
	offset   int64
	basePath string
}

// Name implements billy.File
func (f *billyFile) Name() string {
	return f.name
}

// Write implements billy.File
func (f *billyFile) Write(p []byte) (n int, err error) {
	filePath := filepath.Join(f.basePath, f.name)

	if f.flag&os.O_WRONLY == 0 && f.flag&os.O_RDWR == 0 {
		return 0, os.ErrPermission
	}

	// Create parent directories if needed
	dir := filepath.Dir(filePath)
	if err := f.fs.MkdirAll(dir, 0755); err != nil {
		return 0, err
	}

	// Append to the data
	f.data = append(f.data, p...)
	f.offset += int64(len(p))

	// Write to the filesystem
	if err := f.fs.WriteFile(filePath, f.data, f.perm); err != nil {
		return 0, err
	}

	return len(p), nil
}

// Read implements billy.File
func (f *billyFile) Read(p []byte) (n int, err error) {
	// O_RDONLY is 0, so we only need to check if it's write-only
	if f.flag&os.O_WRONLY != 0 {
		return 0, os.ErrPermission
	}

	// If file doesn't exist and we're not creating it, return EOF
	if len(f.data) == 0 && f.flag&os.O_CREATE == 0 {
		return 0, io.EOF
	}

	if f.offset >= int64(len(f.data)) {
		return 0, io.EOF
	}

	n = copy(p, f.data[f.offset:])
	f.offset += int64(n)
	return n, nil
}

// ReadAt implements billy.File
func (f *billyFile) ReadAt(p []byte, off int64) (n int, err error) {
	// O_RDONLY is 0, so we only need to check if it's write-only
	if f.flag&os.O_WRONLY != 0 {
		return 0, os.ErrPermission
	}

	// If file doesn't exist and we're not creating it, return EOF
	if len(f.data) == 0 && f.flag&os.O_CREATE == 0 {
		return 0, io.EOF
	}

	if off >= int64(len(f.data)) {
		return 0, io.EOF
	}

	n = copy(p, f.data[off:])
	return n, nil
}

// Seek implements billy.File
func (f *billyFile) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = f.offset + offset
	case io.SeekEnd:
		newOffset = int64(len(f.data)) + offset
	default:
		return 0, os.ErrInvalid
	}

	if newOffset < 0 {
		return 0, os.ErrInvalid
	}

	f.offset = newOffset
	return f.offset, nil
}

// Close implements billy.File
func (f *billyFile) Close() error {
	filePath := filepath.Join(f.basePath, f.name)
	// Write any remaining data
	if f.flag&os.O_WRONLY != 0 || f.flag&os.O_RDWR != 0 {
		return f.fs.WriteFile(filePath, f.data, f.perm)
	}
	return nil
}

// Lock implements billy.File
func (f *billyFile) Lock() error {
	// No-op for now
	return nil
}

// Unlock implements billy.File
func (f *billyFile) Unlock() error {
	// No-op for now
	return nil
}

// Truncate implements billy.File
func (f *billyFile) Truncate(size int64) error {
	filePath := filepath.Join(f.basePath, f.name)

	if size < 0 {
		return os.ErrInvalid
	}

	if size > int64(len(f.data)) {
		// Extend the file
		newData := make([]byte, size)
		copy(newData, f.data)
		f.data = newData
	} else {
		// Truncate the file
		f.data = f.data[:size]
	}

	return f.fs.WriteFile(filePath, f.data, f.perm)
}
