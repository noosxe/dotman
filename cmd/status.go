package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/noosxe/dotman/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of the dotfiles",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config
		cfg, err := config.LoadConfig(configPath, fsys)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Open the repository
		repo, err := git.PlainOpen(cfg.DotmanDir)
		if err != nil {
			fmt.Printf("Error opening repository: %v\n", err)
			os.Exit(1)
		}

		// Get the working tree
		worktree, err := repo.Worktree()
		if err != nil {
			fmt.Printf("Error getting worktree: %v\n", err)
			os.Exit(1)
		}

		// Get the status
		status, err := worktree.Status()
		if err != nil {
			fmt.Printf("Error getting status: %v\n", err)
			os.Exit(1)
		}

		// Create a map to store the tree structure
		tree := make(map[string]interface{})

		// Build the tree structure, only including files from data directory
		for file, fileStatus := range status {
			// Skip files not in the data directory
			if !strings.HasPrefix(file, "data/") {
				continue
			}
			// Remove the "data/" prefix for display
			file = strings.TrimPrefix(file, "data/")

			parts := strings.Split(file, string(filepath.Separator))
			current := tree
			for i, part := range parts {
				if i == len(parts)-1 {
					// This is a file - store a copy of the FileStatus struct
					statusCopy := *fileStatus
					current[part] = statusCopy
				} else {
					// This is a directory
					if _, exists := current[part]; !exists {
						current[part] = make(map[string]interface{})
					}
					current = current[part].(map[string]interface{})
				}
			}
		}

		// Print the tree
		fmt.Println("Git Status:")
		fmt.Println("-----------")
		if len(tree) == 0 {
			fmt.Println("Working directory clean")
			return
		}
		printTree(tree, "", true)
	},
}

func printTree(tree map[string]interface{}, prefix string, isLast bool) {
	keys := make([]string, 0, len(tree))
	for k := range tree {
		keys = append(keys, k)
	}

	for i, key := range keys {
		value := tree[key]
		isLastItem := i == len(keys)-1

		// Print the current item
		var currentPrefix string
		if isLast {
			currentPrefix = prefix + "    "
		} else {
			currentPrefix = prefix + "│   "
		}

		var connector string
		if isLastItem {
			connector = "└── "
		} else {
			connector = "├── "
		}

		// Get status symbol
		var status string
		if fileStatus, ok := value.(git.FileStatus); ok {
			// Check both staging and worktree status
			switch {
			case fileStatus.Staging == git.Untracked && fileStatus.Worktree == git.Untracked:
				status = "??"
			case fileStatus.Staging == git.Added:
				status = "A "
			case fileStatus.Staging == git.Modified:
				status = "M "
			case fileStatus.Staging == git.Deleted:
				status = "D "
			case fileStatus.Staging == git.Renamed:
				status = "R "
			case fileStatus.Worktree == git.Modified:
				status = " M"
			case fileStatus.Worktree == git.Deleted:
				status = " D"
			case fileStatus.Worktree == git.Added:
				status = " A"
			default:
				status = "  "
			}
		} else {
			// For directories, show directory icon
			status = "📁"
		}

		fmt.Printf("%s%s%s %s\n", prefix, connector, status, key)

		// If this is a directory, recurse
		if subTree, ok := value.(map[string]interface{}); ok {
			printTree(subTree, currentPrefix, isLastItem)
		}
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
