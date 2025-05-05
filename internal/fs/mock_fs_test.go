package fs

import (
	"io/fs"
	"testing"
)

func TestMockFileSystem_BasicOperations(t *testing.T) {
	mockFS := NewMockFileSystem(nil)

	// Test WriteFile and ReadFile
	testData := []byte("test content")
	err := mockFS.WriteFile("test.txt", testData, 0644)
	if err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	data, err := mockFS.ReadFile("test.txt")
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("ReadFile returned wrong content: got %s, want %s", data, testData)
	}

	// Test Stat
	info, err := mockFS.Stat("test.txt")
	if err != nil {
		t.Errorf("Stat failed: %v", err)
	}
	if info.Name() != "test.txt" {
		t.Errorf("Stat returned wrong name: got %s, want test.txt", info.Name())
	}
	if info.Size() != int64(len(testData)) {
		t.Errorf("Stat returned wrong size: got %d, want %d", info.Size(), len(testData))
	}
	if info.Mode() != 0644 {
		t.Errorf("Stat returned wrong mode: got %v, want 0644", info.Mode())
	}
}

func TestMockFileSystem_DirectoryOperations(t *testing.T) {
	mockFS := NewMockFileSystem(nil)

	// Test MkdirAll
	err := mockFS.MkdirAll("dir/subdir", 0755)
	if err != nil {
		t.Errorf("MkdirAll failed: %v", err)
	}

	// Verify directory exists
	info, err := mockFS.Stat("dir/subdir")
	if err != nil {
		t.Errorf("Stat failed after MkdirAll: %v", err)
	}
	if !info.IsDir() {
		t.Error("MkdirAll did not create a directory")
	}
	if info.Mode() != 0755|fs.ModeDir {
		t.Errorf("MkdirAll created directory with wrong mode: got %v, want %v", info.Mode(), 0755|fs.ModeDir)
	}
}

func TestMockFileSystem_RemoveOperations(t *testing.T) {
	mockFS := NewMockFileSystem(nil)

	// Create test files
	mockFS.WriteFile("file1.txt", []byte("test1"), 0644)
	mockFS.WriteFile("file2.txt", []byte("test2"), 0644)
	mockFS.MkdirAll("dir", 0755)
	mockFS.WriteFile("dir/file3.txt", []byte("test3"), 0644)

	// Test Remove
	err := mockFS.Remove("file1.txt")
	if err != nil {
		t.Errorf("Remove failed: %v", err)
	}
	if _, err := mockFS.Stat("file1.txt"); err == nil {
		t.Error("Remove did not delete the file")
	}

	// Test RemoveAll
	err = mockFS.RemoveAll("dir")
	if err != nil {
		t.Errorf("RemoveAll failed: %v", err)
	}
	if _, err := mockFS.Stat("dir"); err == nil {
		t.Error("RemoveAll did not delete the directory")
	}
	if _, err := mockFS.Stat("dir/file3.txt"); err == nil {
		t.Error("RemoveAll did not delete files in directory")
	}
}

func TestMockFileSystem_Symlink(t *testing.T) {
	mockFS := NewMockFileSystem(nil)

	// Create source file
	testData := []byte("test content")
	mockFS.WriteFile("source.txt", testData, 0644)

	// Test Symlink
	err := mockFS.Symlink("source.txt", "link.txt")
	if err != nil {
		t.Errorf("Symlink failed: %v", err)
	}

	// Verify link exists and points to correct content
	data, err := mockFS.ReadFile("link.txt")
	if err != nil {
		t.Errorf("ReadFile on symlink failed: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("Symlink points to wrong content: got %s, want %s", data, testData)
	}
}

func TestMockFileSystem_PathOperations(t *testing.T) {
	mockFS := NewMockFileSystemWithHome(nil, "/home/test")

	// Test UserHomeDir
	home, err := mockFS.UserHomeDir()
	if err != nil {
		t.Errorf("UserHomeDir failed: %v", err)
	}
	if home != "/home/test" {
		t.Errorf("UserHomeDir returned wrong path: got %s, want /home/test", home)
	}

	// Test Abs
	abs, err := mockFS.Abs("test.txt")
	if err != nil {
		t.Errorf("Abs failed: %v", err)
	}
	if abs != "test.txt" {
		t.Errorf("Abs returned wrong path: got %s, want test.txt", abs)
	}

	// Test Rel
	rel, err := mockFS.Rel("/home/test", "/home/test/file.txt")
	if err != nil {
		t.Errorf("Rel failed: %v", err)
	}
	if rel != "file.txt" {
		t.Errorf("Rel returned wrong path: got %s, want file.txt", rel)
	}
}

func TestMockFileSystem_FileInfo(t *testing.T) {
	mockFS := NewMockFileSystem(nil)
	testData := []byte("test content")
	mockFS.WriteFile("test.txt", testData, 0644)

	info, err := mockFS.Stat("test.txt")
	if err != nil {
		t.Errorf("Stat failed: %v", err)
	}

	// Test FileInfo methods
	if info.Name() != "test.txt" {
		t.Errorf("Name returned wrong value: got %s, want test.txt", info.Name())
	}
	if info.Size() != int64(len(testData)) {
		t.Errorf("Size returned wrong value: got %d, want %d", info.Size(), len(testData))
	}
	if info.Mode() != 0644 {
		t.Errorf("Mode returned wrong value: got %v, want 0644", info.Mode())
	}
	if !info.ModTime().IsZero() {
		t.Error("ModTime should return zero time")
	}
	if info.IsDir() {
		t.Error("IsDir should return false for regular file")
	}
	if info.Sys() != nil {
		t.Error("Sys should return nil")
	}
}
