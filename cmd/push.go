package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/noosxe/dotman/internal/config"
	dotmanfs "github.com/noosxe/dotman/internal/fs"
	"github.com/noosxe/dotman/internal/journal"
	"github.com/spf13/cobra"
)

type pushOperation struct {
	config  *config.Config
	fsys    dotmanfs.FileSystem
	ctx     context.Context
	storage storage.Storer
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push changes to the remote repository",
	Long:  `Push committed changes to the remote repository. This command will push all local commits that haven't been pushed yet.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(configPath, fsys)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create billy filesystem adapter
		billyFs := dotmanfs.NewBillyFileSystem(fsys, cfg.DotmanDir)

		op := &pushOperation{
			fsys:    fsys,
			ctx:     context.Background(),
			config:  cfg,
			storage: filesystem.NewStorage(billyFs, nil),
		}

		return op.run()
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func (op *pushOperation) run() error {
	if err := op.initialize(); err != nil {
		return err
	}

	if err := op.push(); err != nil {
		return err
	}

	return op.complete()
}

func (op *pushOperation) initialize() error {
	// Create journal manager
	jm := journal.NewJournalManager(op.fsys, filepath.Join(op.config.DotmanDir, "journal"))
	if err := jm.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize journal: %w", err)
	}

	// Add journal manager to context
	op.ctx = journal.WithJournalManager(op.ctx, jm)

	// Create journal entry
	entry, err := jm.CreateEntry(journal.OperationTypePush, "", "")
	if err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	// Add entry to context
	op.ctx = journal.WithJournalEntry(op.ctx, entry)

	return nil
}

func (op *pushOperation) push() error {
	// Add push step
	step, err := journal.AddStepToCurrentEntry(op.ctx, journal.StepTypeGit, "Push changes to remote", "", "")
	if err != nil {
		return fmt.Errorf("failed to add push step: %w", err)
	}

	// Start the step
	if err := journal.StartStep(op.ctx, step); err != nil {
		return fmt.Errorf("failed to start step: %w", err)
	}

	// Create billy filesystem adapter
	billyFs := dotmanfs.NewBillyFileSystem(op.fsys, op.config.DotmanDir)

	// Open the repository with our filesystem
	repo, err := git.Open(op.storage, billyFs)
	if err != nil {
		if err := journal.FailEntry(op.ctx, fmt.Errorf("failed to open git repository: %w", err)); err != nil {
			return fmt.Errorf("failed to fail entry: %w", err)
		}
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get the remote
	remote, err := repo.Remote("origin")
	if err != nil {
		if err := journal.FailEntry(op.ctx, fmt.Errorf("failed to get remote: %w", err)); err != nil {
			return fmt.Errorf("failed to fail entry: %w", err)
		}
		return fmt.Errorf("failed to get remote: %w", err)
	}

	// Push changes
	if err := remote.Push(&git.PushOptions{}); err != nil {
		if err := journal.FailEntry(op.ctx, fmt.Errorf("failed to push changes: %w", err)); err != nil {
			return fmt.Errorf("failed to fail entry: %w", err)
		}
		return fmt.Errorf("failed to push changes: %w", err)
	}

	// Complete the step
	if err := journal.CompleteStep(op.ctx, step, "Successfully pushed changes to remote"); err != nil {
		if err := journal.FailEntry(op.ctx, fmt.Errorf("failed to complete step: %w", err)); err != nil {
			return fmt.Errorf("failed to fail entry: %w", err)
		}
		return fmt.Errorf("failed to complete step: %w", err)
	}

	fmt.Println("Successfully pushed changes to remote")
	return nil
}

func (op *pushOperation) complete() error {
	return journal.CompleteEntry(op.ctx)
}
