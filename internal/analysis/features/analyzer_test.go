package features

import (
	"git-history-onboarding/internal/git"
	"regexp"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestNewAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer()
	if analyzer == nil {
		t.Fatal("NewAnalyzer() returned nil")
	}

	// Test that patterns are compiled correctly
	expectedFeatures := []string{
		"Authentication",
		"User Profile",
		"API",
		"Database",
		"UI",
		"Tests",
		"Security",
		"Notifications",
		"Analytics",
		"Cache",
		"Search",
		"Payment",
		"Admin",
		"Monitoring",
		"Logging",
		"Configuration",
		"Scheduling",
		"Caching",
		"Rate Limiting",
		"Documentation",
	}

	for _, feature := range expectedFeatures {
		if patterns, exists := analyzer.featurePatterns[feature]; !exists {
			t.Errorf("Expected feature %s patterns to exist", feature)
		} else if len(patterns) == 0 {
			t.Errorf("Expected feature %s to have patterns", feature)
		}
	}
}

func createTestCommit(hash string, message string, author string, email string, when time.Time, files []string) git.CommitInfo {
	return git.CommitInfo{
		Commit: &object.Commit{
			Hash: plumbing.NewHash(hash),
			Author: object.Signature{
				Name:  author,
				Email: email,
				When:  when,
			},
			Message: message,
		},
		Files: files,
	}
}

func TestAnalyzeCommits(t *testing.T) {
	now := time.Now()
	analyzer := NewAnalyzer()

	// Create test commits
	commits := []git.CommitInfo{
		createTestCommit(
			"abc123def456",
			"feat(auth): implement login",
			"John Doe",
			"john@example.com",
			now.Add(-48*time.Hour),
			[]string{"auth/login.go", "auth/middleware.go"},
		),
		createTestCommit(
			"def456abc789",
			"fix(auth): fix session handling",
			"Jane Smith",
			"jane@example.com",
			now.Add(-24*time.Hour),
			[]string{"auth/session.go"},
		),
		createTestCommit(
			"789abc123def",
			"feat(api): add user endpoints",
			"John Doe",
			"john@example.com",
			now,
			[]string{"api/user.go", "api/routes.go"},
		),
	}

	// Analyze commits
	features := analyzer.AnalyzeCommits(commits)

	// Test Authentication feature
	t.Run("Authentication Feature", func(t *testing.T) {
		auth, exists := features["Authentication"]
		assert.True(t, exists, "Authentication feature should exist")
		if exists {
			assert.Len(t, auth.Commits, 2, "Should have 2 commits")
			assert.Len(t, auth.Bugs, 1, "Should have 1 bug")
			assert.Equal(t, "john@example.com", getHighestOwner(auth.Owners))
		}
	})

	// Test API feature
	t.Run("API Feature", func(t *testing.T) {
		api, exists := features["API"]
		assert.True(t, exists, "API feature should exist")
		if exists {
			assert.Len(t, api.Commits, 1, "Should have 1 commit")
			assert.Empty(t, api.Bugs, "Should have no bugs")
			assert.Equal(t, "john@example.com", getHighestOwner(api.Owners))
		}
	})

	// Test feature detection from file paths
	t.Run("File Path Detection", func(t *testing.T) {
		commit := createTestCommit(
			"test123",
			"update database schema",
			"Test User",
			"test@example.com",
			now,
			[]string{"db/schema.go", "db/migrations/001.sql"},
		)

		features := analyzer.AnalyzeCommits([]git.CommitInfo{commit})
		db, exists := features["Database"]
		assert.True(t, exists, "Database feature should be detected from file path")
		if exists {
			assert.Len(t, db.Commits, 1)
		}
	})

	// Test conventional commit parsing
	t.Run("Conventional Commit Parsing", func(t *testing.T) {
		commit := createTestCommit(
			"test456",
			"feat(ui): add new component\n\nBREAKING CHANGE: new API",
			"Test User",
			"test@example.com",
			now,
			[]string{"ui/components/new.tsx"},
		)

		features := analyzer.AnalyzeCommits([]git.CommitInfo{commit})
		ui, exists := features["UI"]
		assert.True(t, exists, "UI feature should be detected from conventional commit")
		if exists {
			assert.Len(t, ui.Commits, 1)
		}
	})
}

// Helper function to get the owner with highest ownership percentage
func getHighestOwner(owners map[string]float64) string {
	var highestOwner string
	var highestPercentage float64

	for owner, percentage := range owners {
		if percentage > highestPercentage {
			highestPercentage = percentage
			highestOwner = owner
		}
	}

	return highestOwner
}

func TestFeaturePatternMatching(t *testing.T) {
	analyzer := NewAnalyzer()

	testCases := []struct {
		name     string
		path     string
		feature  string
		expected bool
	}{
		{"Auth File", "src/auth/login.go", "Authentication", true},
		{"API File", "api/v1/users.go", "API", true},
		{"Database File", "internal/db/schema.go", "Database", true},
		{"Test File", "pkg/service/service_test.go", "Tests", true},
		{"Random File", "README.md", "Authentication", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patterns := analyzer.featurePatterns[tc.feature]
			matched := analyzer.matchesFeature(tc.path, patterns)
			assert.Equal(t, tc.expected, matched)
		})
	}
}

func TestParseConventionalCommit(t *testing.T) {
	testCases := []struct {
		name     string
		message  string
		expected *ConventionalCommit
	}{
		{
			name:    "Simple feat commit",
			message: "feat: add new feature",
			expected: &ConventionalCommit{
				Type:        "feat",
				Description: "add new feature",
				Breaking:    false,
			},
		},
		{
			name:    "Commit with scope",
			message: "fix(auth): resolve login issue",
			expected: &ConventionalCommit{
				Type:        "fix",
				Scope:       "auth",
				Description: "resolve login issue",
				Breaking:    false,
			},
		},
		{
			name:    "Breaking change with !",
			message: "feat(api)!: redesign API",
			expected: &ConventionalCommit{
				Type:        "feat",
				Scope:       "api",
				Description: "redesign API",
				Breaking:    true,
			},
		},
		{
			name: "Commit with body and breaking change",
			message: `feat(db): add new database
			
BREAKING CHANGE: This changes the database schema
Migration required`,
			expected: &ConventionalCommit{
				Type:        "feat",
				Scope:       "db",
				Description: "add new database",
				Body:        "Migration required",
				Breaking:    true,
			},
		},
		{
			name:     "Invalid format",
			message:  "just a regular commit message",
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseConventionalCommit(tc.message)

			if tc.expected == nil {
				if result != nil {
					t.Errorf("Expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if result.Type != tc.expected.Type {
				t.Errorf("Expected type %s, got %s", tc.expected.Type, result.Type)
			}

			if result.Scope != tc.expected.Scope {
				t.Errorf("Expected scope %s, got %s", tc.expected.Scope, result.Scope)
			}

			if result.Description != tc.expected.Description {
				t.Errorf("Expected description %s, got %s", 
					tc.expected.Description, result.Description)
			}

			if result.Breaking != tc.expected.Breaking {
				t.Errorf("Expected breaking %v, got %v", 
					tc.expected.Breaking, result.Breaking)
			}
		})
	}
}

func TestMatchesFeature(t *testing.T) {
	analyzer := NewAnalyzer()

	testCases := []struct {
		name     string
		input    string
		patterns []*regexp.Regexp
		expected bool
	}{
		{
			name:     "Auth file match",
			input:    "src/auth/login.go",
			patterns: analyzer.featurePatterns["Authentication"],
			expected: true,
		},
		{
			name:     "API file match",
			input:    "api/endpoint.go",
			patterns: analyzer.featurePatterns["API"],
			expected: true,
		},
		{
			name:     "No match",
			input:    "random/file.go",
			patterns: analyzer.featurePatterns["Authentication"],
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.matchesFeature(tc.input, tc.patterns)
			if result != tc.expected {
				t.Errorf("Expected %v for input %s", tc.expected, tc.input)
			}
		})
	}
}

func TestIsBugFix(t *testing.T) {
	analyzer := NewAnalyzer()

	testCases := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "Fix commit",
			message:  "fix: resolve issue",
			expected: true,
		},
		{
			name:     "Bug in message",
			message:  "Fixed a bug in authentication",
			expected: true,
		},
		{
			name:     "Regular commit",
			message:  "Add new feature",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.isBugFix(tc.message)
			if result != tc.expected {
				t.Errorf("Expected %v for message %s", tc.expected, tc.message)
			}
		})
	}
} 