package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/noosxe/dotman/internal/config"
	dotmanfs "github.com/noosxe/dotman/internal/fs"
	"github.com/noosxe/dotman/internal/journal"
	"github.com/spf13/cobra"
)

// addOperation represents the state of an add operation
type addOperation struct {
	path   string
	config *config.Config
	fsys   dotmanfs.FileSystem
	ctx    context.Context
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new dotfile to the dotman repository",
	Long:  `Add a new dotfile to the dotman repository by specifying the path to the file or the directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")

		op := &addOperation{
			path: path,
			fsys: fsys,
		}

		if err := op.run(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully added and verified %s to dotman repository\n", path)
	},
}

func (op *addOperation) run() error {
	if err := op.initialize(); err != nil {
		return err
	}

	if err := op.verifySource(); err != nil {
		return err
	}

	if err := op.copyAndVerify(); err != nil {
		return err
	}

	if err := op.createSymlink(); err != nil {
		return err
	}

	if err := op.gitAdd(); err != nil {
		return err
	}

	return op.complete()
}

func (op *addOperation) initialize() error {
	// Load config
	cfg, err := config.LoadConfig(configPath, op.fsys)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}
	op.config = cfg

	// Get user's home directory using fsys
	homeDir, err := op.fsys.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %v", err)
	}

	// Check if the path is within the home directory
	absPath, err := op.fsys.Abs(op.path)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %v", err)
	}

	// Get relative path from home directory
	relPath, err := op.fsys.Rel(homeDir, absPath)
	if err != nil {
		return fmt.Errorf("error getting relative path: %v", err)
	}

	// If the path is not within home directory, return error
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path must be within user's home directory")
	}

	// Initialize journal manager
	jm := journal.NewJournalManager(op.fsys, filepath.Join(cfg.DotmanDir, "journal"))
	if err := jm.Initialize(); err != nil {
		return fmt.Errorf("error initializing journal: %v", err)
	}

	// Create journal entry with the relative path as target
	entry, err := jm.CreateEntry(journal.OperationTypeAdd, op.path, relPath)
	if err != nil {
		return fmt.Errorf("error creating journal entry: %v", err)
	}

	// Add journal manager and entry to context
	op.ctx = journal.WithJournalManager(context.Background(), jm)
	op.ctx = journal.WithJournalEntry(op.ctx, entry)

	return nil
}

func (op *addOperation) verifySource() error {
	// Create verification step
	step, err := journal.AddStepToCurrentEntry(op.ctx, journal.StepTypeVerify, "Verify source path exists", op.path, "")
	if err != nil {
		return err
	}

	// Start verification step
	if err := journal.StartStep(op.ctx, step); err != nil {
		return err
	}

	// Perform verification
	info, err := op.fsys.Stat(op.path)
	if err != nil {
		// Fail the entire entry
		if err := journal.FailEntry(op.ctx, err); err != nil {
			return err
		}
		return fmt.Errorf("source path does not exist: %v", err)
	}

	// Complete verification step
	details := fmt.Sprintf("Path exists and is a %s", map[bool]string{true: "directory", false: "file"}[info.IsDir()])
	if err := journal.CompleteStep(op.ctx, step, details); err != nil {
		return err
	}

	return nil
}

func (op *addOperation) copyAndVerify() error {
	info, _ := op.fsys.Stat(op.path)
	entry, _ := journal.GetJournalEntry(op.ctx)
	targetPath := filepath.Join(op.config.DotmanDir, "data", entry.Target)

	if info.IsDir() {
		return op.copyAndVerifyDirectory(targetPath)
	}
	return op.copyAndVerifyFile(targetPath)
}

func (op *addOperation) copyAndVerifyDirectory(targetPath string) error {
	// Add directory copy step
	step, err := journal.AddStepToCurrentEntry(op.ctx, journal.StepTypeCopy, "Copy directory contents", op.path, targetPath)
	if err != nil {
		return err
	}

	// Start copy step
	if err := journal.StartStep(op.ctx, step); err != nil {
		return err
	}

	// Copy directory
	if err := copyDir(op.path, targetPath, op.fsys); err != nil {
		if err := journal.FailEntry(op.ctx, err); err != nil {
			return err
		}
		return fmt.Errorf("error copying directory: %v", err)
	}

	// Complete copy step
	if err := journal.CompleteStep(op.ctx, step, "Successfully copied all directory contents"); err != nil {
		return err
	}

	// Add verification step
	verifyStep, err := journal.AddStepToCurrentEntry(op.ctx, journal.StepTypeVerify, "Verify directory copy", op.path, targetPath)
	if err != nil {
		return err
	}

	// Start verification step
	if err := journal.StartStep(op.ctx, verifyStep); err != nil {
		return err
	}

	// Verify directory copy
	if err := verifyDirCopy(op.path, targetPath, op.fsys); err != nil {
		if err := journal.FailEntry(op.ctx, err); err != nil {
			return err
		}
		return fmt.Errorf("error verifying directory copy: %v", err)
	}

	// Complete verification step
	if err := journal.CompleteStep(op.ctx, verifyStep, "Successfully verified all directory contents match"); err != nil {
		return err
	}

	return nil
}

func (op *addOperation) copyAndVerifyFile(targetPath string) error {
	// Add file copy step
	step, err := journal.AddStepToCurrentEntry(op.ctx, journal.StepTypeCopy, "Copy file contents", op.path, targetPath)
	if err != nil {
		return err
	}

	// Start copy step
	if err := journal.StartStep(op.ctx, step); err != nil {
		return err
	}

	// Copy file
	if err := copyFile(op.path, targetPath, op.fsys); err != nil {
		if err := journal.FailEntry(op.ctx, err); err != nil {
			return err
		}
		return fmt.Errorf("error copying file: %v", err)
	}

	// Complete copy step
	if err := journal.CompleteStep(op.ctx, step, "Successfully copied file contents"); err != nil {
		return err
	}

	// Add verification step
	verifyStep, err := journal.AddStepToCurrentEntry(op.ctx, journal.StepTypeVerify, "Verify file copy", op.path, targetPath)
	if err != nil {
		return err
	}

	// Start verification step
	if err := journal.StartStep(op.ctx, verifyStep); err != nil {
		return err
	}

	// Verify file copy
	if err := verifyFileCopy(op.path, targetPath, op.fsys); err != nil {
		if err := journal.FailEntry(op.ctx, err); err != nil {
			return err
		}
		return fmt.Errorf("error verifying file copy: %v", err)
	}

	// Complete verification step
	if err := journal.CompleteStep(op.ctx, verifyStep, "Successfully verified file contents match"); err != nil {
		return err
	}

	return nil
}

func (op *addOperation) createSymlink() error {
	entry, _ := journal.GetJournalEntry(op.ctx)
	targetPath := filepath.Join(op.config.DotmanDir, "data", entry.Target)

	// Add symlink step
	step, err := journal.AddStepToCurrentEntry(op.ctx, journal.StepTypeSymlink, "Create symlink", op.path, targetPath)
	if err != nil {
		return err
	}

	// Start symlink step
	if err := journal.StartStep(op.ctx, step); err != nil {
		return err
	}

	// Remove original file/directory
	if err := op.fsys.RemoveAll(op.path); err != nil {
		if err := journal.FailEntry(op.ctx, err); err != nil {
			return err
		}
		return fmt.Errorf("error removing original file/directory: %v", err)
	}

	// Create symlink
	if err := op.fsys.Symlink(targetPath, op.path); err != nil {
		if err := journal.FailEntry(op.ctx, err); err != nil {
			return err
		}
		return fmt.Errorf("error creating symlink: %v", err)
	}

	// Complete symlink step
	if err := journal.CompleteStep(op.ctx, step, "Successfully created symlink"); err != nil {
		return err
	}

	return nil
}

func (op *addOperation) gitAdd() error {
	// Add git add step
	step, err := journal.AddStepToCurrentEntry(op.ctx, journal.StepTypeGit, "Add file to git", op.path, "")
	if err != nil {
		return err
	}

	// Start git add step
	if err := journal.StartStep(op.ctx, step); err != nil {
		return err
	}

	// Open the repository
	repo, err := git.PlainOpen(op.config.DotmanDir)
	if err != nil {
		if err := journal.FailEntry(op.ctx, err); err != nil {
			return err
		}
		return fmt.Errorf("error opening repository: %v", err)
	}

	// Get the worktree
	worktree, err := repo.Worktree()
	if err != nil {
		if err := journal.FailEntry(op.ctx, err); err != nil {
			return err
		}
		return fmt.Errorf("error getting worktree: %v", err)
	}

	// Add the file to git using the relative path
	entry, _ := journal.GetJournalEntry(op.ctx)
	targetPath := filepath.Join("data", entry.Target)
	fmt.Println("Adding file to git:", targetPath)
	if _, err := worktree.Add(targetPath); err != nil {
		if err := journal.FailEntry(op.ctx, err); err != nil {
			return err
		}
		return fmt.Errorf("error adding file to git: %v", err)
	}

	// Complete git add step
	if err := journal.CompleteStep(op.ctx, step, "Successfully added file to git"); err != nil {
		return err
	}

	return nil
}

func (op *addOperation) complete() error {
	return journal.CompleteEntry(op.ctx)
}

func copyFile(src, dst string, fsys dotmanfs.FileSystem) error {
	file, err := fsys.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	data := make([]byte, info.Size())
	if _, err := file.Read(data); err != nil {
		return err
	}

	return fsys.WriteFile(dst, data, info.Mode())
}

func verifyFileCopy(src, dst string, fsys dotmanfs.FileSystem) error {
	srcFile, err := fsys.Open(src)
	if err != nil {
		return fmt.Errorf("error reading source file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := fsys.Open(dst)
	if err != nil {
		return fmt.Errorf("error reading destination file: %v", err)
	}
	defer dstFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("error getting source file info: %v", err)
	}

	dstInfo, err := dstFile.Stat()
	if err != nil {
		return fmt.Errorf("error getting destination file info: %v", err)
	}

	if srcInfo.Size() != dstInfo.Size() {
		return fmt.Errorf("file sizes differ: source=%d bytes, destination=%d bytes", srcInfo.Size(), dstInfo.Size())
	}

	srcData := make([]byte, srcInfo.Size())
	dstData := make([]byte, dstInfo.Size())

	if _, err := srcFile.Read(srcData); err != nil {
		return fmt.Errorf("error reading source file content: %v", err)
	}

	if _, err := dstFile.Read(dstData); err != nil {
		return fmt.Errorf("error reading destination file content: %v", err)
	}

	for i := range srcData {
		if srcData[i] != dstData[i] {
			return fmt.Errorf("file contents differ at byte %d", i)
		}
	}

	return nil
}

func copyDir(src, dst string, fsys dotmanfs.FileSystem) error {
	// Create destination directory
	if err := fsys.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read source directory
	dir, err := fsys.Open(src)
	if err != nil {
		return err
	}
	defer dir.Close()

	entries, err := dir.(fs.ReadDirFile).ReadDir(-1)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath, fsys); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath, fsys); err != nil {
				return err
			}
		}
	}

	return nil
}

func verifyDirCopy(src, dst string, fsys dotmanfs.FileSystem) error {
	srcDir, err := fsys.Open(src)
	if err != nil {
		return fmt.Errorf("error reading source directory: %v", err)
	}
	defer srcDir.Close()

	dstDir, err := fsys.Open(dst)
	if err != nil {
		return fmt.Errorf("error reading destination directory: %v", err)
	}
	defer dstDir.Close()

	srcEntries, err := srcDir.(fs.ReadDirFile).ReadDir(-1)
	if err != nil {
		return fmt.Errorf("error reading source directory entries: %v", err)
	}

	dstEntries, err := dstDir.(fs.ReadDirFile).ReadDir(-1)
	if err != nil {
		return fmt.Errorf("error reading destination directory entries: %v", err)
	}

	if len(srcEntries) != len(dstEntries) {
		return fmt.Errorf("directory contents differ: source has %d entries, destination has %d entries", len(srcEntries), len(dstEntries))
	}

	for i, srcEntry := range srcEntries {
		dstEntry := dstEntries[i]
		if srcEntry.Name() != dstEntry.Name() {
			return fmt.Errorf("directory entries differ: source has %s, destination has %s", srcEntry.Name(), dstEntry.Name())
		}

		srcPath := filepath.Join(src, srcEntry.Name())
		dstPath := filepath.Join(dst, dstEntry.Name())

		if srcEntry.IsDir() {
			if !dstEntry.IsDir() {
				return fmt.Errorf("entry type mismatch: %s is a directory in source but not in destination", srcEntry.Name())
			}
			if err := verifyDirCopy(srcPath, dstPath, fsys); err != nil {
				return err
			}
		} else {
			if dstEntry.IsDir() {
				return fmt.Errorf("entry type mismatch: %s is a file in source but a directory in destination", srcEntry.Name())
			}
			if err := verifyFileCopy(srcPath, dstPath, fsys); err != nil {
				return fmt.Errorf("error verifying file %s: %v", srcEntry.Name(), err)
			}
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringP("path", "p", "", "path to the dotfile")
	addCmd.MarkFlagRequired("path")
}
