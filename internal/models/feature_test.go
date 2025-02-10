package models

import (
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"git-history-onboarding/internal/git"
)

func TestFeature(t *testing.T) {
	// Create test data
	now := time.Now()
	testCommit := &object.Commit{
		Hash: plumbing.NewHash("abcdef1234567890"),
		Author: object.Signature{
			Name:  "Test Author",
			Email: "test@example.com",
			When:  now,
		},
		Message: "test commit",
	}

	commitInfo := git.CommitInfo{
		Commit: testCommit,
		Files:  []string{"test/file1.go", "test/file2.go"},
	}

	// Create a feature
	feature := &Feature{
		Name:         "Test Feature",
		CreatedAt:    now,
		LastUpdated:  now,
		Owners:       make(map[string]float64),
		BackupOwners: make(map[string]float64),
		Commits:      []git.CommitInfo{commitInfo},
		Bugs:         []Bug{},
	}

	// Test feature properties
	t.Run("Feature Properties", func(t *testing.T) {
		assert.Equal(t, "Test Feature", feature.Name)
		assert.Equal(t, now, feature.CreatedAt)
		assert.Equal(t, now, feature.LastUpdated)
		assert.Len(t, feature.Commits, 1)
		assert.Empty(t, feature.Bugs)
	})

	// Test adding a bug
	t.Run("Add Bug", func(t *testing.T) {
		bug := Bug{
			CommitHash:    testCommit.Hash.String(),
			Description:   "Test bug fix",
			ReportedAt:   now,
			FixedAt:      now,
			AffectedFiles: []string{"test/file1.go"},
			AuthorEmail:   "test@example.com",
		}

		feature.Bugs = append(feature.Bugs, bug)
		assert.Len(t, feature.Bugs, 1)
		assert.Equal(t, bug, feature.Bugs[0])
	})

	// Test ownership calculation
	t.Run("Ownership Calculation", func(t *testing.T) {
		feature.Owners["test@example.com"] = 1.0
		assert.Equal(t, 1.0, feature.Owners["test@example.com"])
	})
}

func TestBug(t *testing.T) {
	now := time.Now()
	bug := Bug{
		CommitHash:    "abcdef1234567890",
		Description:   "Test bug fix",
		ReportedAt:   now,
		FixedAt:      now,
		AffectedFiles: []string{"test/file1.go"},
		AuthorEmail:   "test@example.com",
	}

	assert.Equal(t, "abcdef1234567890", bug.CommitHash)
	assert.Equal(t, "Test bug fix", bug.Description)
	assert.Equal(t, now, bug.ReportedAt)
	assert.Equal(t, now, bug.FixedAt)
	assert.Equal(t, []string{"test/file1.go"}, bug.AffectedFiles)
	assert.Equal(t, "test@example.com", bug.AuthorEmail)
}

// TestBugLifecycle tests the complete lifecycle of a bug
func TestBugLifecycle(t *testing.T) {
	now := time.Now()
	reportTime := now.Add(-48 * time.Hour)
	fixTime := now.Add(-24 * time.Hour)

	bug := Bug{
		CommitHash:    "fix789",
		Description:   "Critical bug fix",
		ReportedAt:    reportTime,
		FixedAt:       fixTime,
		AffectedFiles: []string{"critical/file.go", "utils/helper.go"},
		AuthorEmail:   "developer@example.com",
	}

	// Test bug timeline
	if !bug.ReportedAt.Before(bug.FixedAt) {
		t.Error("Expected ReportedAt to be before FixedAt")
	}

	// Test resolution time
	resolutionTime := bug.FixedAt.Sub(bug.ReportedAt)
	expectedResolutionTime := 24 * time.Hour
	if resolutionTime != expectedResolutionTime {
		t.Errorf("Expected resolution time %v, got %v", 
			expectedResolutionTime, resolutionTime)
	}

	// Test affected files
	if len(bug.AffectedFiles) != 2 {
		t.Errorf("Expected 2 affected files, got %d", len(bug.AffectedFiles))
	}
} 