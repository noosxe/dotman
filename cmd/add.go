package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/noosxe/dotman/internal/config"
	dotmanfs "github.com/noosxe/dotman/internal/fs"
	"github.com/noosxe/dotman/internal/journal"
	"github.com/spf13/cobra"
)

// addOperation represents the state of an add operation
type addOperation struct {
	path           string
	config         *config.Config
	journalManager *journal.JournalManager
	fsys           dotmanfs.FileSystem
	entry          *journal.JournalEntry
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

	return op.complete()
}

func (op *addOperation) initialize() error {
	// Load config
	cfg, err := config.LoadConfig(configPath, op.fsys)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}
	op.config = cfg

	// Initialize journal manager
	op.journalManager = journal.NewJournalManager(op.fsys, filepath.Join(cfg.DotmanDir, "journal"))
	if err := op.journalManager.Initialize(); err != nil {
		return fmt.Errorf("error initializing journal: %v", err)
	}

	// Create journal entry
	entry, err := op.journalManager.CreateEntry("add", op.path, filepath.Join(cfg.DotmanDir, filepath.Base(op.path)))
	if err != nil {
		return fmt.Errorf("error creating journal entry: %v", err)
	}
	op.entry = entry

	return nil
}

func (op *addOperation) verifySource() error {
	// Check if path exists
	info, err := op.fsys.Stat(op.path)
	if err != nil {
		op.entry.Steps = append(op.entry.Steps, journal.Step{
			Type:        "verify",
			Status:      "failed",
			Error:       err.Error(),
			Description: "Verify source path exists",
			Source:      op.path,
			StartTime:   time.Now(),
			EndTime:     time.Now(),
		})
		op.journalManager.UpdateEntry(op.entry)
		op.journalManager.MoveEntry(op.entry, "failed")
		return fmt.Errorf("source path does not exist: %v", err)
	}

	// Add verification step
	verifyStart := time.Now()
	op.entry.Steps = append(op.entry.Steps, journal.Step{
		Type:        "verify",
		Status:      "completed",
		Description: "Verify source path exists",
		Source:      op.path,
		Details:     fmt.Sprintf("Path exists and is a %s", map[bool]string{true: "directory", false: "file"}[info.IsDir()]),
		StartTime:   verifyStart,
		EndTime:     time.Now(),
	})
	if err := op.journalManager.UpdateEntry(op.entry); err != nil {
		return fmt.Errorf("error updating journal: %v", err)
	}

	return nil
}

func (op *addOperation) copyAndVerify() error {
	info, _ := op.fsys.Stat(op.path)
	targetPath := filepath.Join(op.config.DotmanDir, filepath.Base(op.path))

	if info.IsDir() {
		return op.copyAndVerifyDirectory(targetPath)
	}
	return op.copyAndVerifyFile(targetPath)
}

func (op *addOperation) copyAndVerifyDirectory(targetPath string) error {
	// Add directory copy step
	copyStart := time.Now()
	op.entry.Steps = append(op.entry.Steps, journal.Step{
		Type:        "copy",
		Status:      "in-progress",
		Description: "Copy directory contents",
		Source:      op.path,
		Target:      targetPath,
		StartTime:   copyStart,
	})
	if err := op.journalManager.UpdateEntry(op.entry); err != nil {
		return fmt.Errorf("error updating journal: %v", err)
	}

	// Copy directory
	if err := copyDir(op.path, targetPath, op.fsys); err != nil {
		op.entry.Steps[len(op.entry.Steps)-1].Status = "failed"
		op.entry.Steps[len(op.entry.Steps)-1].Error = err.Error()
		op.entry.Steps[len(op.entry.Steps)-1].EndTime = time.Now()
		op.journalManager.UpdateEntry(op.entry)
		op.journalManager.MoveEntry(op.entry, "failed")
		return fmt.Errorf("error copying directory: %v", err)
	}

	// Update copy step
	op.entry.Steps[len(op.entry.Steps)-1].Status = "completed"
	op.entry.Steps[len(op.entry.Steps)-1].Details = "Successfully copied all directory contents"
	op.entry.Steps[len(op.entry.Steps)-1].EndTime = time.Now()
	if err := op.journalManager.UpdateEntry(op.entry); err != nil {
		return fmt.Errorf("error updating journal: %v", err)
	}

	// Add verification step
	verifyStart := time.Now()
	op.entry.Steps = append(op.entry.Steps, journal.Step{
		Type:        "verify",
		Status:      "in-progress",
		Description: "Verify directory copy",
		Source:      op.path,
		Target:      targetPath,
		StartTime:   verifyStart,
	})
	if err := op.journalManager.UpdateEntry(op.entry); err != nil {
		return fmt.Errorf("error updating journal: %v", err)
	}

	// Verify directory copy
	if err := verifyDirCopy(op.path, targetPath, op.fsys); err != nil {
		op.entry.Steps[len(op.entry.Steps)-1].Status = "failed"
		op.entry.Steps[len(op.entry.Steps)-1].Error = err.Error()
		op.entry.Steps[len(op.entry.Steps)-1].EndTime = time.Now()
		op.journalManager.UpdateEntry(op.entry)
		op.journalManager.MoveEntry(op.entry, "failed")
		return fmt.Errorf("error verifying directory copy: %v", err)
	}

	// Update verification step
	op.entry.Steps[len(op.entry.Steps)-1].Status = "completed"
	op.entry.Steps[len(op.entry.Steps)-1].Details = "Successfully verified all directory contents match"
	op.entry.Steps[len(op.entry.Steps)-1].EndTime = time.Now()

	return nil
}

func (op *addOperation) copyAndVerifyFile(targetPath string) error {
	// Add file copy step
	copyStart := time.Now()
	op.entry.Steps = append(op.entry.Steps, journal.Step{
		Type:        "copy",
		Status:      "in-progress",
		Description: "Copy file contents",
		Source:      op.path,
		Target:      targetPath,
		StartTime:   copyStart,
	})
	if err := op.journalManager.UpdateEntry(op.entry); err != nil {
		return fmt.Errorf("error updating journal: %v", err)
	}

	// Copy file
	if err := copyFile(op.path, targetPath, op.fsys); err != nil {
		op.entry.Steps[len(op.entry.Steps)-1].Status = "failed"
		op.entry.Steps[len(op.entry.Steps)-1].Error = err.Error()
		op.entry.Steps[len(op.entry.Steps)-1].EndTime = time.Now()
		op.journalManager.UpdateEntry(op.entry)
		op.journalManager.MoveEntry(op.entry, "failed")
		return fmt.Errorf("error copying file: %v", err)
	}

	// Update copy step
	op.entry.Steps[len(op.entry.Steps)-1].Status = "completed"
	op.entry.Steps[len(op.entry.Steps)-1].Details = "Successfully copied file contents"
	op.entry.Steps[len(op.entry.Steps)-1].EndTime = time.Now()
	if err := op.journalManager.UpdateEntry(op.entry); err != nil {
		return fmt.Errorf("error updating journal: %v", err)
	}

	// Add verification step
	verifyStart := time.Now()
	op.entry.Steps = append(op.entry.Steps, journal.Step{
		Type:        "verify",
		Status:      "in-progress",
		Description: "Verify file copy",
		Source:      op.path,
		Target:      targetPath,
		StartTime:   verifyStart,
	})
	if err := op.journalManager.UpdateEntry(op.entry); err != nil {
		return fmt.Errorf("error updating journal: %v", err)
	}

	// Verify file copy
	if err := verifyFileCopy(op.path, targetPath, op.fsys); err != nil {
		op.entry.Steps[len(op.entry.Steps)-1].Status = "failed"
		op.entry.Steps[len(op.entry.Steps)-1].Error = err.Error()
		op.entry.Steps[len(op.entry.Steps)-1].EndTime = time.Now()
		op.journalManager.UpdateEntry(op.entry)
		op.journalManager.MoveEntry(op.entry, "failed")
		return fmt.Errorf("error verifying file copy: %v", err)
	}

	// Update verification step
	op.entry.Steps[len(op.entry.Steps)-1].Status = "completed"
	op.entry.Steps[len(op.entry.Steps)-1].Details = "Successfully verified file contents match"
	op.entry.Steps[len(op.entry.Steps)-1].EndTime = time.Now()

	return nil
}

func (op *addOperation) complete() error {
	// Move entry to completed state
	if err := op.journalManager.UpdateEntry(op.entry); err != nil {
		return fmt.Errorf("error updating journal: %v", err)
	}
	if err := op.journalManager.MoveEntry(op.entry, "completed"); err != nil {
		return fmt.Errorf("error moving journal entry: %v", err)
	}

	return nil
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
