package ownership

import (
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"git-history-onboarding/internal/git"
	"git-history-onboarding/internal/models"
)

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

func TestNewAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer(0.2, 0.1)
	assert.NotNil(t, analyzer)
	assert.Equal(t, 0.2, analyzer.PrimaryThreshold)
	assert.Equal(t, 0.1, analyzer.BackupThreshold)
}

func TestUpdateFeatureOwnership(t *testing.T) {
	// Use lower thresholds for testing
	analyzer := NewAnalyzer(0.4, 0.2)  // 40% for primary, 20% for backup
	now := time.Now()

	// Create test commits - 2 from John (67%) and 1 from Jane (33%)
	commits := []git.CommitInfo{
		createTestCommit(
			"abc123",
			"feat: first commit",
			"John Doe",
			"john@example.com",
			now.Add(-48*time.Hour),
			[]string{"file1.go"},
		),
		createTestCommit(
			"def456",
			"fix: second commit",
			"Jane Smith",
			"jane@example.com",
			now.Add(-24*time.Hour),
			[]string{"file2.go"},
		),
		createTestCommit(
			"ghi789",
			"feat: third commit",
			"John Doe",
			"john@example.com",
			now,
			[]string{"file3.go"},
		),
	}

	feature := &models.Feature{
		Name:         "Test Feature",
		Commits:      commits,
		Owners:       make(map[string]float64),
		BackupOwners: make(map[string]float64),
	}

	analyzer.UpdateFeatureOwnership(feature)

	// John should be primary owner (67% > 40%)
	assert.Contains(t, feature.Owners, "john@example.com")
	assert.InDelta(t, 0.67, feature.Owners["john@example.com"], 0.01)

	// Jane should be backup owner (33% > 20%)
	assert.Contains(t, feature.BackupOwners, "jane@example.com")
	assert.InDelta(t, 0.33, feature.BackupOwners["jane@example.com"], 0.01)
}

func TestAnalyzeOwnership(t *testing.T) {
	now := time.Now()
	analyzer := NewAnalyzer(0.4, 0.2)  // 40% for primary, 20% for backup

	tests := []struct {
		name          string
		commits       []git.CommitInfo
		wantOwners   map[string]float64
		wantBackups  map[string]float64
	}{
		{
			name: "Single owner",
			commits: []git.CommitInfo{
				createTestCommit("abc123", "test commit", "John Doe", "john@example.com", now, []string{"file1.go"}),
				createTestCommit("def456", "test commit", "John Doe", "john@example.com", now.Add(time.Hour), []string{"file2.go"}),
			},
			wantOwners: map[string]float64{
				"john@example.com": 1.0,
			},
			wantBackups: map[string]float64{},
		},
		{
			name: "Multiple owners - all backup",
			commits: []git.CommitInfo{
				createTestCommit("abc123", "test commit", "John Doe", "john@example.com", now, []string{"file1.go"}),
				createTestCommit("def456", "test commit", "Jane Smith", "jane@example.com", now.Add(time.Hour), []string{"file2.go"}),
				createTestCommit("ghi789", "test commit", "Bob Wilson", "bob@example.com", now.Add(2*time.Hour), []string{"file3.go"}),
			},
			wantOwners: map[string]float64{},
			wantBackups: map[string]float64{
				"john@example.com": 0.333,
				"jane@example.com": 0.333,
				"bob@example.com":  0.333,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owners, backups := analyzer.AnalyzeOwnership(tt.commits)
			
			assert.Equal(t, len(tt.wantOwners), len(owners))
			assert.Equal(t, len(tt.wantBackups), len(backups))

			for email, percentage := range tt.wantOwners {
				assert.InDelta(t, percentage, owners[email], 0.01)
			}

			for email, percentage := range tt.wantBackups {
				assert.InDelta(t, percentage, backups[email], 0.01)
			}
		})
	}
}

func TestGetTopOwners(t *testing.T) {
	now := time.Now()
	analyzer := NewAnalyzer(0.2, 0.1)

	commits := []git.CommitInfo{
		createTestCommit("abc123", "test commit", "John Doe", "john@example.com", now, []string{"file1.go"}),
		createTestCommit("def456", "test commit", "John Doe", "john@example.com", now.Add(time.Hour), []string{"file2.go"}),
		createTestCommit("ghi789", "test commit", "Jane Smith", "jane@example.com", now.Add(2*time.Hour), []string{"file3.go"}),
		createTestCommit("jkl012", "test commit", "Bob Wilson", "bob@example.com", now.Add(3*time.Hour), []string{"file4.go"}),
	}

	tests := []struct {
		name string
		n    int
		want []string
	}{
		{
			name: "Top 2 owners",
			n:    2,
			want: []string{"john@example.com", "jane@example.com"},
		},
		{
			name: "All owners",
			n:    3,
			want: []string{"john@example.com", "jane@example.com", "bob@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.GetTopOwners(commits, tt.n)
			assert.Equal(t, tt.want, got)
		})
	}
} 