package ownership

import (
	"sort"
	"git-history-onboarding/internal/git"
	"git-history-onboarding/internal/models"
)

type Analyzer struct {
	PrimaryThreshold float64
	BackupThreshold float64
}

func NewAnalyzer(primaryThreshold, backupThreshold float64) *Analyzer {
	if primaryThreshold == 0 {
		primaryThreshold = 0.2 // 20% default
	}
	if backupThreshold == 0 {
		backupThreshold = 0.1 // 10% default
	}
	return &Analyzer{
		PrimaryThreshold: primaryThreshold,
		BackupThreshold: backupThreshold,
	}
}

func (a *Analyzer) UpdateFeatureOwnership(feature *models.Feature) {
	primary, backup := a.AnalyzeOwnership(feature.Commits)
	feature.Owners = primary
	feature.BackupOwners = backup
}

func (a *Analyzer) AnalyzeOwnership(commits []git.CommitInfo) (map[string]float64, map[string]float64) {
	if len(commits) == 0 {
		return make(map[string]float64), make(map[string]float64)
	}

	// Count commits per author
	commitCounts := make(map[string]int)
	for _, commit := range commits {
		commitCounts[commit.Commit.Author.Email]++
	}

	// Calculate percentages
	totalCommits := float64(len(commits))
	owners := make(map[string]float64)
	backups := make(map[string]float64)

	// First pass: identify primary owners
	for email, count := range commitCounts {
		percentage := float64(count) / totalCommits
		if percentage >= a.PrimaryThreshold {
			owners[email] = percentage
		}
	}

	// Second pass: if no primary owners, treat all as backup owners
	if len(owners) == 0 {
		for email, count := range commitCounts {
			percentage := float64(count) / totalCommits
			if percentage >= a.BackupThreshold {
				backups[email] = percentage
			}
		}
	} else {
		// If we have primary owners, remaining contributors become backup owners
		for email, count := range commitCounts {
			if _, isPrimary := owners[email]; !isPrimary {
				percentage := float64(count) / totalCommits
				if percentage >= a.BackupThreshold {
					backups[email] = percentage
				}
			}
		}
	}

	return owners, backups
}

func (a *Analyzer) GetTopOwners(commits []git.CommitInfo, n int) []string {
	if n <= 0 || len(commits) == 0 {
		return nil
	}

	// Count commits per author
	commitCounts := make(map[string]int)
	for _, commit := range commits {
		commitCounts[commit.Commit.Author.Email]++
	}

	// Convert to slice for sorting
	type authorCount struct {
		email string
		count int
	}
	authors := make([]authorCount, 0, len(commitCounts))
	for email, count := range commitCounts {
		authors = append(authors, authorCount{email, count})
	}

	// Sort by count descending
	sort.Slice(authors, func(i, j int) bool {
		return authors[i].count > authors[j].count
	})

	// Get top N authors
	result := make([]string, 0, n)
	for i := 0; i < n && i < len(authors); i++ {
		result = append(result, authors[i].email)
	}

	return result
} 