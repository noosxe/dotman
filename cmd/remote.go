package cmd

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/noosxe/dotman/internal/config"
	"github.com/spf13/cobra"
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage git remote repository",
	Long:  `Manage the git remote repository used for syncing dotfiles.`,
}

var remoteShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current remote URL",
	Long:  `Display the URL of the git remote repository used for syncing dotfiles.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config
		cfg, err := config.LoadConfig(configPath, fsys)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Open the repository
		repo, err := git.PlainOpen(cfg.DotmanDir)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Get the remote
		remote, err := repo.Remote("origin")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Get the URL
		urls := remote.Config().URLs
		if len(urls) == 0 {
			fmt.Println("No remote URL configured")
			os.Exit(1)
		}

		fmt.Println("Remote URL:", urls[0])
	},
}

var remoteSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the remote URL",
	Long:  `Set the URL of the git remote repository used for syncing dotfiles.`,
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		if url == "" {
			fmt.Println("Error: URL is required")
			os.Exit(1)
		}

		// Load config
		cfg, err := config.LoadConfig(configPath, fsys)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Open the repository
		repo, err := git.PlainOpen(cfg.DotmanDir)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Remove existing remote if it exists
		_, err = repo.Remote("origin")
		if err == nil {
			if err := repo.DeleteRemote("origin"); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}

		// Create new remote
		_, err = repo.CreateRemote(&gitconfig.RemoteConfig{
			Name: "origin",
			URLs: []string{url},
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully set remote URL to: %s\n", url)
	},
}

func init() {
	rootCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(remoteShowCmd)
	remoteCmd.AddCommand(remoteSetCmd)

	remoteSetCmd.Flags().StringP("url", "u", "", "URL of the git remote repository")
	remoteSetCmd.MarkFlagRequired("url")
}
