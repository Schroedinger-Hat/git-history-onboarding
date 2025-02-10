package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"git-history-onboarding/internal/git"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "git-analyzer",
		Short: "Analyze git repository history",
		Run:   analyze,
	}

	rootCmd.Flags().StringP("repo", "r", "", "Repository URL to analyze")
	rootCmd.MarkFlagRequired("repo")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func analyze(cmd *cobra.Command, args []string) {
	repoURL, _ := cmd.Flags().GetString("repo")
	
	ctx := context.Background()
	repo, err := git.Clone(ctx, repoURL)
	if err != nil {
		log.Fatalf("Failed to clone repository: %v", err)
	}

	commits, err := repo.GetCommitHistory()
	if err != nil {
		log.Fatalf("Failed to get commit history: %v", err)
	}

	fmt.Printf("Found %d commits\n", len(commits))
	
	// Print commit information and changed files
	for _, commit := range commits {
		fmt.Printf("\nCommit: %s\n", commit.Commit.Hash)
		fmt.Printf("Author: %s\n", commit.Commit.Author.Name)
		fmt.Printf("Message: %s\n", commit.Commit.Message)
		fmt.Printf("Changed files:\n")
		for _, file := range commit.Files {
			fmt.Printf("  - %s\n", file)
		}
	}
} 