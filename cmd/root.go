package cmd

import (
	"github.com/spf13/cobra"
)

var (
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dotman",
	Short: "A simple and efficient dotfiles manager",
	Long: `dotman is a command-line tool designed to help you manage your dotfiles
across different machines. It provides a simple way to track, version,
and deploy your configuration files.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
