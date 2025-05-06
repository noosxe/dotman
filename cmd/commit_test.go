package cmd

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/noosxe/dotman/internal/config"
	dotmanfs "github.com/noosxe/dotman/internal/fs"
	"github.com/noosxe/dotman/internal/journal"
	"github.com/noosxe/dotman/internal/testutil"
)

func TestCommitOperation(t *testing.T) {
	// Create mock filesystem with dotman structure
	fsys, dotmanDir := testutil.NewMockFSWithDotman()

	// Create config
	cfg := &config.Config{
		DotmanDir: dotmanDir,
	}
	configPath := filepath.Join(testutil.TestHomeDir, ".dotconfig")
	if err := config.SaveConfig(configPath, cfg, fsys); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

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

	// Create a sample file for initial commit
	sampleFile := filepath.Join(dotmanDir, "data", "sample.txt")
	if err := fsys.WriteFile(sampleFile, []byte("sample content"), 0644); err != nil {
		t.Fatalf("failed to create sample file: %v", err)
	}

	// Add the sample file to git
	if _, err := worktree.Add("data/sample.txt"); err != nil {
		t.Fatalf("failed to add sample file: %v", err)
	}

	// Create initial commit
	if _, err := worktree.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "dotman",
			Email: "dotman@localhost",
		},
	}); err != nil {
		t.Fatalf("failed to create initial commit: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(dotmanDir, "data", "test.txt")
	if err := fsys.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Add test file to git
	if _, err := worktree.Add("data/test.txt"); err != nil {
		t.Fatalf("failed to add test file: %v", err)
	}

	// Create journal manager
	jm := journal.NewJournalManager(fsys, filepath.Join(dotmanDir, "journal"))
	if err := jm.Initialize(); err != nil {
		t.Fatalf("failed to initialize journal: %v", err)
	}

	// Create context with journal manager
	ctx := context.Background()
	ctx = journal.WithJournalManager(ctx, jm)

	// Create commit operation
	op := &commitOperation{
		message: "test commit",
		fsys:    fsys,
		ctx:     ctx,
		config:  cfg,
		storage: memStorage,
	}

	// Execute commit
	if err := op.run(); err != nil {
		t.Fatalf("failed to execute commit: %v", err)
	}

	// Verify git commit message
	commitIter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		t.Fatalf("failed to get commit log: %v", err)
	}

	commit, err := commitIter.Next()
	if err != nil {
		t.Fatalf("failed to get commit: %v", err)
	}

	if commit.Message != "test commit" {
		t.Fatalf("expected commit message 'test commit', got '%s'", commit.Message)
	}

	// Verify journal entry was created
	entries, err := jm.ListEntries(string(journal.EntryStateCompleted))
	if err != nil {
		t.Fatalf("failed to get journal entries: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("no journal entries found")
	}

	lastEntry := entries[len(entries)-1]
	testutil.VerifyEntryWithSteps(t, lastEntry, journal.OperationTypeCommit, journal.EntryStateCompleted, 1)

	step := lastEntry.Steps[0]
	testutil.VerifyStep(t, step, journal.StepTypeGit, journal.StepStatusCompleted, "test commit")
}
