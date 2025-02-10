package git

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Setup code here
	code := m.Run()
	// Cleanup code here
	os.Exit(code)
}

func TestClone(t *testing.T) {
	// Test repository URL - using a small public repo
	testRepoURL := "https://github.com/golang/example.git"
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test cloning
	repo, err := Clone(ctx, testRepoURL)
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	// Test getting commit history
	commits, err := repo.GetCommitHistory()
	assert.NoError(t, err)
	assert.NotEmpty(t, commits)

	// Verify commit information
	for _, commit := range commits {
		assert.NotNil(t, commit.Commit)
		assert.NotEmpty(t, commit.Commit.Hash.String())
		assert.NotEmpty(t, commit.Commit.Author.Name)
		assert.NotEmpty(t, commit.Commit.Author.Email)
		assert.False(t, commit.Commit.Author.When.IsZero())
	}
}

func TestCloneInvalidRepo(t *testing.T) {
	ctx := context.Background()
	
	// Test with invalid repository URL
	repo, err := Clone(ctx, "https://github.com/not-exists/not-exists.git")
	assert.Error(t, err)
	assert.Nil(t, repo)
}

// TestCloneWithCanceledContext tests the behavior when the context is canceled
func TestCloneWithCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	repo, err := Clone(ctx, "https://github.com/golang/example.git")
	
	// Check that we got an error
	assert.Error(t, err)
	assert.Nil(t, repo)
	
	// Check that the error is context-related
	assert.True(t, 
		strings.Contains(err.Error(), "context canceled") || 
		err == context.Canceled,
		"Expected context cancellation error, got: %v", err)
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) > len(substr) && s[len(s)-len(substr):] == substr
} 