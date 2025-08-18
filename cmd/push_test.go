package cmd

import (
	"testing"

	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/noosxe/dotman/internal/journal"
	"github.com/noosxe/dotman/internal/testutil"
)

func TestPushOperation(t *testing.T) {
	// Create mock filesystem with dotman structure
	fsys, dotmanDir, err := testutil.NewMockFSWithDotman()
	if err != nil {
		t.Fatalf("failed to create mock filesystem: %v", err)
	}
	defer fsys.CleanUp()

	// Setup test config
	cfg := testutil.SetupTestConfig(t, fsys, dotmanDir)

	// Setup git repository
	repo, worktree, storage := testutil.SetupTestGitRepo(t, fsys, dotmanDir)

	// Create initial commit with sample file
	testutil.CreateTestFileAndCommit(t, fsys, worktree, dotmanDir, "data/sample.txt", "sample content")

	_ = testutil.SetupBareRepo(t, fsys, "home/remote")

	repo.CreateRemote(&gitconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{fsys.RealPath("home/remote")},
	})

	// Setup journal manager
	jm := testutil.SetupJournalManager(t, fsys, dotmanDir)

	// Setup context with journal
	ctx := testutil.SetupContextWithJournal(t, jm, journal.OperationTypePush, "", "")

	// Create push operation
	op := &pushOperation{
		fsys:    fsys,
		ctx:     ctx,
		config:  cfg,
		storage: storage,
	}

	// Execute push
	if err := op.run(); err != nil {
		t.Fatalf("failed to execute push: %v\n\n%v", err, fsys.DumpTree())
	}

	// Verify journal entry was created
	testutil.VerifyJournalEntryCount(t, jm, journal.EntryStateCompleted, 1)

	// Get the last entry and verify its details
	entries, err := jm.ListEntries(journal.EntryStateCompleted)
	if err != nil {
		t.Fatalf("failed to get journal entries: %v", err)
	}

	lastEntry := entries[0]
	testutil.VerifyEntryWithSteps(t, lastEntry, journal.OperationTypePush, journal.EntryStateCompleted, 1)

	step := lastEntry.Steps[0]
	testutil.VerifyStep(t, step, journal.StepTypeGit, journal.StepStatusCompleted, "Push changes to remote")
}
