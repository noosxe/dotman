package cmd

import (
	"testing"

	"github.com/noosxe/dotman/internal/journal"
	"github.com/noosxe/dotman/internal/testutil"
)

func TestCommitOperation(t *testing.T) {
	// Create mock filesystem with dotman structure
	fsys, dotmanDir := testutil.NewMockFSWithDotman()

	// Setup test config
	cfg := testutil.SetupTestConfig(t, fsys, dotmanDir)

	// Setup git repository
	repo, worktree, storage := testutil.SetupTestGitRepo(t, fsys, dotmanDir)

	// Create initial commit with sample file
	testutil.CreateTestFileAndCommit(t, fsys, worktree, dotmanDir, "data/sample.txt", "sample content")

	// Create a test file and add it to git (without committing)
	testutil.CreateTestFileAndAdd(t, fsys, worktree, dotmanDir, "data/test.txt", "test content")

	// Setup journal manager
	jm := testutil.SetupJournalManager(t, fsys, dotmanDir)

	// Setup context with journal
	ctx := testutil.SetupContextWithJournal(t, jm, journal.OperationTypeCommit, "", "")

	// Create commit operation
	op := &commitOperation{
		message: "test commit",
		fsys:    fsys,
		ctx:     ctx,
		config:  cfg,
		storage: storage,
	}

	// Execute commit
	if err := op.run(); err != nil {
		t.Fatalf("failed to execute commit: %v", err)
	}

	// Verify git commit message
	testutil.VerifyLastCommit(t, repo, "test commit")

	// Verify journal entry was created
	testutil.VerifyJournalEntryCount(t, jm, journal.EntryStateCompleted, 1)

	// Get the last entry and verify its details
	entries, err := jm.ListEntries(journal.EntryStateCompleted)
	if err != nil {
		t.Fatalf("failed to get journal entries: %v", err)
	}

	lastEntry := entries[0]
	testutil.VerifyEntryWithSteps(t, lastEntry, journal.OperationTypeCommit, journal.EntryStateCompleted, 1)

	step := lastEntry.Steps[0]
	testutil.VerifyStep(t, step, journal.StepTypeGit, journal.StepStatusCompleted, "test commit")
}
