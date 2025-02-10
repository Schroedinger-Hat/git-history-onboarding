package models

import (
	"time"
	"git-history-onboarding/internal/git"
)

// Feature represents a software feature and its associated metadata
type Feature struct {
	Name         string
	Path         string
	Commits      []git.CommitInfo
	Owners       map[string]float64  // email -> ownership percentage
	BackupOwners map[string]float64  // email -> ownership percentage
	CreatedAt    time.Time
	LastUpdated  time.Time
	Bugs         []Bug
}

// Bug represents a bug fix in the codebase
type Bug struct {
	CommitHash    string
	Description   string
	ReportedAt    time.Time
	FixedAt       time.Time
	AffectedFiles []string
	AuthorEmail   string    // Changed from Author to AuthorEmail
}

// FeatureDetectionConfig holds configuration for feature detection
type FeatureDetectionConfig struct {
	// Patterns to identify features from commit messages
	FeaturePrefixes []string // e.g., "feat:", "feature:"
	BugPrefixes     []string // e.g., "fix:", "bug:"
	
	// Directory-based feature detection
	FeaturePaths    map[string]string // e.g., "auth/" -> "Authentication"
} 