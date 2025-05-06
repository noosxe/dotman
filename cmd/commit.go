package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/noosxe/dotman/internal/config"
	dotmanfs "github.com/noosxe/dotman/internal/fs"
	"github.com/noosxe/dotman/internal/journal"
	"github.com/spf13/cobra"
)

type commitOperation struct {
	// mandatory fields
	config *config.Config
	fsys   dotmanfs.FileSystem
	ctx    context.Context

	// additional fields required for commit operation
	message string
	storage storage.Storer
}

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit changes to the journal",
	Long: `Commit changes to the journal with a descriptive message.
This command will record the current state of tracked files in the journal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		message, _ := cmd.Flags().GetString("message")
		if message == "" {
			return fmt.Errorf("commit message is required")
		}

		cfg, err := config.LoadConfig(configPath, fsys)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create billy filesystem adapter
		billyFs := dotmanfs.NewBillyFileSystem(fsys, cfg.DotmanDir)

		op := &commitOperation{
			message: message,
			fsys:    fsys,
			ctx:     context.Background(),
			config:  cfg,
			storage: filesystem.NewStorage(billyFs, nil),
		}

		return op.run()
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().StringP("message", "m", "", "commit message")
}

func (op *commitOperation) run() error {
	if err := op.initialize(); err != nil {
		return err
	}

	if err := op.commit(); err != nil {
		return err
	}

	return op.complete()
}

func (op *commitOperation) initialize() error {
	// Create journal manager
	jm := journal.NewJournalManager(op.fsys, filepath.Join(op.config.DotmanDir, "journal"))
	if err := jm.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize journal: %w", err)
	}

	// Add journal manager to context
	op.ctx = journal.WithJournalManager(op.ctx, jm)

	// Create journal entry
	entry, err := jm.CreateEntry(journal.OperationTypeCommit, "", "")
	if err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	// Add entry to context
	op.ctx = journal.WithJournalEntry(op.ctx, entry)

	return nil
}

func (op *commitOperation) commit() error {
	// Add commit step
	step, err := journal.AddStepToCurrentEntry(op.ctx, journal.StepTypeGit, op.message, "", "")
	if err != nil {
		return fmt.Errorf("failed to add commit step: %w", err)
	}

	// Start the step
	if err := journal.StartStep(op.ctx, step); err != nil {
		return fmt.Errorf("failed to start step: %w", err)
	}

	// Create billy filesystem adapter
	billyFs := dotmanfs.NewBillyFileSystem(op.fsys, op.config.DotmanDir)

	// Open git repository with our filesystem
	repo, err := git.Open(op.storage, billyFs)
	if err != nil {
		if err := journal.FailEntry(op.ctx, fmt.Errorf("failed to open git repository: %w", err)); err != nil {
			return fmt.Errorf("failed to fail entry: %w", err)
		}
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		if err := journal.FailEntry(op.ctx, fmt.Errorf("failed to get worktree: %w", err)); err != nil {
			return fmt.Errorf("failed to fail entry: %w", err)
		}
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add all changes
	if err := worktree.AddGlob("."); err != nil {
		if err := journal.FailEntry(op.ctx, fmt.Errorf("failed to add changes: %w", err)); err != nil {
			return fmt.Errorf("failed to fail entry: %w", err)
		}
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Get author info from git config
	gitCfg, err := repo.ConfigScoped(gitconfig.GlobalScope)
	if err != nil {
		if err := journal.FailEntry(op.ctx, fmt.Errorf("failed to get git config: %w", err)); err != nil {
			return fmt.Errorf("failed to fail entry: %w", err)
		}
		return fmt.Errorf("failed to get git config: %w", err)
	}

	// Commit changes
	commit, err := worktree.Commit(op.message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  gitCfg.User.Name,
			Email: gitCfg.User.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		if err := journal.FailEntry(op.ctx, fmt.Errorf("failed to commit changes: %w", err)); err != nil {
			return fmt.Errorf("failed to fail entry: %w", err)
		}
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Get commit hash
	commitObj, err := repo.CommitObject(commit)
	if err != nil {
		if err := journal.FailEntry(op.ctx, fmt.Errorf("failed to get commit object: %w", err)); err != nil {
			return fmt.Errorf("failed to fail entry: %w", err)
		}
		return fmt.Errorf("failed to get commit object: %w", err)
	}

	// Complete the step with commit hash
	if err := journal.CompleteStep(op.ctx, step, fmt.Sprintf("Committed changes with hash: %s", commitObj.Hash.String())); err != nil {
		if err := journal.FailEntry(op.ctx, fmt.Errorf("failed to complete step: %w", err)); err != nil {
			return fmt.Errorf("failed to fail entry: %w", err)
		}
		return fmt.Errorf("failed to complete step: %w", err)
	}

	fmt.Printf("Changes committed successfully with hash: %s\n", commitObj.Hash.String())
	return nil
}

func (op *commitOperation) complete() error {
	return journal.CompleteEntry(op.ctx)
}
