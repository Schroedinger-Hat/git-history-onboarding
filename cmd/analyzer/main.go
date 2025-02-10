package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"git-history-onboarding/internal/git"
	"git-history-onboarding/internal/analysis/features"
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
	
	// Analyze features
	analyzer := features.NewAnalyzer()
	featureAnalysis := analyzer.AnalyzeCommits(commits)

	// Print feature analysis
	fmt.Println("\nFeature Analysis:")
	for name, feature := range featureAnalysis {
		fmt.Printf("\nFeature: %s\n", name)
		fmt.Printf("Created: %s\n", feature.CreatedAt.Format("2006-01-02"))
		fmt.Printf("Last Updated: %s\n", feature.LastUpdated.Format("2006-01-02"))
		
		fmt.Println("Primary Owners:")
		for email, percentage := range feature.Owners {
			fmt.Printf("  - %s (%.1f%%)\n", email, percentage*100)
		}
		
		fmt.Println("Backup Owners:")
		for email, percentage := range feature.BackupOwners {
			fmt.Printf("  - %s (%.1f%%)\n", email, percentage*100)
		}
		
		fmt.Printf("Number of Commits: %d\n", len(feature.Commits))
		fmt.Printf("Number of Bugs: %d\n", len(feature.Bugs))
		
		if len(feature.Bugs) > 0 {
			fmt.Println("Bug History:")
			for _, bug := range feature.Bugs {
				fmt.Printf("  - [%s] by %s: %s\n", 
					bug.FixedAt.Format("2006-01-02"),
					bug.AuthorEmail,
					bug.Description)
			}
		}
	}
} 