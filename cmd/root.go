package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	dotmanfs "github.com/noosxe/dotman/internal/fs"
	"github.com/spf13/cobra"
)

var (
	configPath string
	verbose    bool
	fsys       = dotmanfs.NewOSFileSystem()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dotman",
	Short: "A dotfile manager",
	Long: `dotman is a CLI tool for managing dotfiles.
It helps you track, version control, and sync your dotfiles across different machines.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Get default config path
	home, err := os.UserHomeDir()
	if err != nil {
		home = "~"
	}
	defaultConfigPath := filepath.Join(home, ".dotconfig")

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", defaultConfigPath, "path to config file (default is $HOME/.dotconfig)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
