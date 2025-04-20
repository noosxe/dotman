package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	force bool
)

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
		// TODO: Implement initialization logic
		fmt.Println("dotman initialized")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Local flags for init command
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "force initialization even if directory is not empty")
}
