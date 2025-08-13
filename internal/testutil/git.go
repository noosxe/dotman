package testutil

import (
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/noosxe/dotman/internal/config"
	dotmanfs "github.com/noosxe/dotman/internal/fs"
)

// SetupTestGitRepo creates a git repository in the given directory with an initial commit
func SetupTestGitRepo(t *testing.T, fsys dotmanfs.FileSystem, dotmanDir string) (*git.Repository, *git.Worktree, storage.Storer) {
	// Create billy filesystem adapter
	billyFs := dotmanfs.NewBillyFileSystem(fsys, dotmanDir)

	// Create memory storage
	memStorage := memory.NewStorage()

	// Initialize git repository with memory storage
	repo, err := git.InitWithOptions(memStorage, billyFs, git.InitOptions{
		DefaultBranch: "refs/heads/main",
	})
	if err != nil {
		t.Fatalf("failed to initialize git repository: %v", err)
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	return repo, worktree, memStorage
}

// CreateTestFileAndAdd creates a test file and adds it to git without committing
func CreateTestFileAndAdd(t *testing.T, fsys dotmanfs.FileSystem, worktree *git.Worktree, dotmanDir, filePath, content string) {
	// Create the file
	fullPath := filepath.Join(dotmanDir, filePath)
	if err := fsys.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Add file to git
	if _, err := worktree.Add(filePath); err != nil {
		t.Fatalf("failed to add test file: %v", err)
	}
}

// CreateTestFileAndCommit creates a test file, adds it to git, and commits it
func CreateTestFileAndCommit(t *testing.T, fsys dotmanfs.FileSystem, worktree *git.Worktree, dotmanDir, filePath, content string) {
	// Create and add the file
	CreateTestFileAndAdd(t, fsys, worktree, dotmanDir, filePath, content)

	// Create commit
	if _, err := worktree.Commit("test commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "dotman",
			Email: "dotman@localhost",
		},
	}); err != nil {
		t.Fatalf("failed to create commit: %v", err)
	}
}

// VerifyLastCommit verifies that the last commit in the repository has the expected message
func VerifyLastCommit(t *testing.T, repo *git.Repository, expectedMessage string) {
	commitIter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		t.Fatalf("failed to get commit log: %v", err)
	}

	commit, err := commitIter.Next()
	if err != nil {
		t.Fatalf("failed to get commit: %v", err)
	}

	if commit.Message != expectedMessage {
		t.Fatalf("expected commit message '%s', got '%s'", expectedMessage, commit.Message)
	}
}

// SetupTestConfig creates and saves a test configuration
func SetupTestConfig(t *testing.T, fsys dotmanfs.FileSystem, dotmanDir string) *config.Config {
	cfg := &config.Config{
		DotmanDir: dotmanDir,
	}
	configPath := filepath.Join(TestHomeDir, ".dotconfig")
	if err := config.SaveConfig(configPath, cfg, fsys); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	return cfg
}

func SetupBareRepo(t *testing.T, fsys dotmanfs.FileSystem, storage storage.Storer, dir string) *git.Repository {
	// fsys.MkdirAll(dir, 0755)

	// Create billy filesystem adapter
	billyFs := dotmanfs.NewBillyFileSystem(fsys, dir)

	memStorage := memory.NewStorage()

	repo, err := git.InitWithOptions(memStorage, billyFs, git.InitOptions{
		DefaultBranch: "refs/heads/main",
	})

	if err != nil {
		t.Fatalf("failed to initialize git repository: %v", err)
	}
	return repo
}
