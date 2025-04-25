package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/noosxe/dotman/internal/config"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new dotfile to the dotman repository",
	Long:  `Add a new dotfile to the dotman repository by specifying the path to the file or the directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")

		// Load config
		cfg, err := config.LoadConfig(configPath, fsys)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Check if path exists
		info, err := os.Stat(path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Create target path in dotman directory
		targetPath := filepath.Join(cfg.DotmanDir, filepath.Base(path))

		// Copy file or directory
		if info.IsDir() {
			// Copy directory
			if err := copyDir(path, targetPath); err != nil {
				fmt.Printf("Error copying directory: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Copy file
			if err := copyFile(path, targetPath); err != nil {
				fmt.Printf("Error copying file: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("Successfully added %s to dotman repository\n", path)
	},
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
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
