package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new dotfile to the dotman repository",
	Long:  `Add a new dotfile to the dotman repository by specifying the path to the file or the directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Adding dotfile...")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringP("path", "p", "", "path to the dotfile")
	addCmd.MarkFlagRequired("path")
}
