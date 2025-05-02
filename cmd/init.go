package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	dotmanconfig "github.com/noosxe/dotman/internal/config"
	"github.com/spf13/cobra"
)

var (
	force bool
	dir   string
)

// isDotmanDir checks if a directory is a dotman directory by checking for .manfile
func isDotmanDir(path string) bool {
	manfile := filepath.Join(path, ".manfile")
	_, err := os.Stat(manfile)
	return err == nil
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize dotman in the current directory",
	Long: `Initialize dotman in the current directory by creating necessary
configuration files and directory structure.`,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			fmt.Println("Initializing dotman...")
		}

		// Check if directory exists
		info, err := os.Stat(dir)
		if err == nil {
			if !info.IsDir() {
				fmt.Printf("Error: %s exists but is not a directory\n", dir)
				os.Exit(1)
			}

			if isDotmanDir(dir) && !force {
				fmt.Printf("Error: %s is already a dotman directory. Use --force to overwrite\n", dir)
				os.Exit(1)
			}

			if !force {
				fmt.Printf("Error: %s already exists. Use --force to overwrite\n", dir)
				os.Exit(1)
			}

			if verbose {
				fmt.Printf("Force flag used, deleting existing directory: %s\n", dir)
			}

			// Remove existing directory if force is true
			if err := os.RemoveAll(dir); err != nil {
				fmt.Printf("Error removing directory: %v\n", err)
				os.Exit(1)
			}

			if verbose {
				fmt.Printf("Directory deleted successfully: %s\n", dir)
			}
		}

		// Create directory
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			os.Exit(1)
		}

		// Create .manfile
		manfile := filepath.Join(dir, ".manfile")
		if err := os.WriteFile(manfile, []byte("{}"), 0644); err != nil {
			fmt.Printf("Error creating .manfile: %v\n", err)
			os.Exit(1)
		}

		// Create .gitignore
		gitignore := filepath.Join(dir, ".gitignore")
		gitignoreContent := `# dotman specific
journal/
config.json

# Common patterns
*.swp
*.swo
*~
.DS_Store
`
		if err := os.WriteFile(gitignore, []byte(gitignoreContent), 0644); err != nil {
			fmt.Printf("Error creating .gitignore: %v\n", err)
			os.Exit(1)
		}

		repo, err := git.PlainInitWithOptions(dir, &git.PlainInitOptions{
			Bare: false,
			InitOptions: git.InitOptions{
				DefaultBranch: "refs/heads/main",
			},
		})

		if err != nil {
			fmt.Printf("Error initializing git repository: %v\n", err)
			os.Exit(1)
		}

		if verbose {
			fmt.Printf("Git repository initialized successfully: %s\n", dir)
		}

		wt, err := repo.Worktree()
		if err != nil {
			fmt.Printf("Error getting worktree: %v\n", err)
			os.Exit(1)
		}

		wt.Add(".manfile")
		wt.Add(".gitignore")

		// Get author info from git config
		gitCfg, err := repo.ConfigScoped(gitconfig.GlobalScope)
		if err != nil {
			fmt.Printf("Error getting git config: %v\n", err)
			os.Exit(1)
		}

		if _, err := wt.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  gitCfg.User.Name,
				Email: gitCfg.User.Email,
				When:  time.Now(),
			},
		}); err != nil {
			fmt.Printf("Error committing .manfile: %v\n", err)
			os.Exit(1)
		}

		// Save dotman directory to config
		cfg, err := dotmanconfig.LoadConfig(configPath, fsys)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		cfg.DotmanDir = dir
		if err := dotmanconfig.SaveConfig(configPath, cfg, fsys); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("dotman initialized in %s\n", dir)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	home, err := os.UserHomeDir()
	if err != nil {
		home = "~"
	}
	defaultDir := filepath.Join(home, ".dotman")

	// Local flags for init command
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "force initialization even if directory is not empty")
	initCmd.Flags().StringVarP(&dir, "dir", "d", defaultDir, "directory to initialize dotman in")
}
