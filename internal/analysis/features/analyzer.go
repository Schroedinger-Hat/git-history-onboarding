package features

import (
	"path/filepath"
	"regexp"
	"strings"

	"git-history-onboarding/internal/git"
	"git-history-onboarding/internal/models"
	"git-history-onboarding/internal/analysis/ownership"
)

type Analyzer struct {
	featurePatterns map[string][]*regexp.Regexp
	ownershipAnalyzer *ownership.Analyzer
}

type ConventionalCommit struct {
	Type        string
	Scope       string
	Description string
	Body        string
	Breaking    bool
}

func NewAnalyzer() *Analyzer {
	patterns := map[string][]string{
		"Authentication": {`auth`, `login`, `oauth`, `sign[ui][pn]`, `signout`},
		"User Profile": {`profile`, `user[-_]?(?:profile|settings|management|dashboard)?`, `account`},
		"API": {`api(?:[-_](?:gateway|client|server|docs|documentation))?`, `graphql`, `rest`},
		"Database": {`db`, `database`, `storage`, `sql`, `nosql`, `orm`, `migration`},
		"UI": {`ui`, `interface`, `component`, `theme`, `style`, `css`, `html`, `javascript`, `react`, `vue`, `angular`, `svelte`, `tailwind`, `bootstrap`},
		"Tests": {`test`, `spec`, `_test\.go$`},
		"Security": {`auth`, `security`, `authentication`, `authorization`, `encrypt(?:ion)?`, `hash(?:ing)?`, `password`, `token`, `jwt`, `api[-_](?:key|token|secret)`},
        "Notifications":{`notification`, `notifier`, `notify`, `alert`, `toast`, `snackbar`},
        "Analytics": {`analytics`, `tracking`, `telemetry`, `metrics`, `stats`, `logger`, `logging`},
        "Cache": {`cache`, `memcached`, `redis`, `caching`},
        "Search": {`search`, `indexing`, `fulltext`, `autocomplete`, `filter`, `sort`},
        "Payment": {`payment`, `billing`, `subscription`, `invoice`, `purchase`},
        "Admin": {`admin`, `dashboard`, `management`, `control`, `panel`},
        "Monitoring": {`monitor`, `observe`, `stats`, `metrics`, `logging`, `tracing`},
        "Logging": {`log`, `logger`, `logging`, `syslog`, `journald`},
        "Configuration": {`config`, `configuration`, `settings`, `properties`, `properties`},
        "Scheduling": {`schedule`, `scheduler`, `cron`, `job`, `task`},
        "Caching": {`cache`, `memcached`, `redis`, `caching`},
        "Rate Limiting": {`rate`, `limit`, `limiter`, `throttle`},
        "Documentation": {`docs`, `documentation`, `readme`, `changelog`, `release`, `upgrade`, `migration`},
	}

	compiledPatterns := make(map[string][]*regexp.Regexp)
	for feature, patternList := range patterns {
		compiledPatterns[feature] = make([]*regexp.Regexp, 0, len(patternList))
		for _, pattern := range patternList {
			regex := regexp.MustCompile(`(?i)` + pattern)  // (?i) makes it case-insensitive
			compiledPatterns[feature] = append(compiledPatterns[feature], regex)
		}
	}

	return &Analyzer{
		featurePatterns: compiledPatterns,
		ownershipAnalyzer: ownership.NewAnalyzer(0.2, 0.1),
	}
}

func (a *Analyzer) AnalyzeCommits(commits []git.CommitInfo) map[string]*models.Feature {
	features := make(map[string]*models.Feature)

	// Initialize features
	for name := range a.featurePatterns {
		features[name] = &models.Feature{
			Name:         name,
			Owners:       make(map[string]float64),
			BackupOwners: make(map[string]float64),
			Commits:      make([]git.CommitInfo, 0),
			Bugs:         make([]models.Bug, 0),
		}
	}

	// Analyze each commit
	for _, commit := range commits {
		a.processCommit(commit, features)
	}

	// Update ownership for each feature
	for _, feature := range features {
		a.ownershipAnalyzer.UpdateFeatureOwnership(feature)
	}

	return features
}

func (a *Analyzer) processCommit(commit git.CommitInfo, features map[string]*models.Feature) {
	// Parse conventional commit format
	conventionalCommit := parseConventionalCommit(commit.Commit.Message)
	
	for featureName, patterns := range a.featurePatterns {
		feature := features[featureName]
		matchFound := false

		// Check conventional commit scope if available
		if conventionalCommit != nil && conventionalCommit.Scope != "" {
			matchFound = a.matchesFeature(conventionalCommit.Scope, patterns)
		}

		// If no match found in scope, check the description and body
		if !matchFound && conventionalCommit != nil {
			matchFound = a.matchesFeature(conventionalCommit.Description, patterns) ||
						a.matchesFeature(conventionalCommit.Body, patterns)
		}

		// If still no match, check the files
		if !matchFound {
			for _, file := range commit.Files {
				if a.matchesFeature(file, patterns) {
					matchFound = true
					break
				}
			}
		}

		if matchFound {
			// Update feature information
			if feature.CreatedAt.IsZero() || commit.Commit.Author.When.Before(feature.CreatedAt) {
				feature.CreatedAt = commit.Commit.Author.When
			}
			if commit.Commit.Author.When.After(feature.LastUpdated) {
				feature.LastUpdated = commit.Commit.Author.When
			}

			feature.Commits = append(feature.Commits, commit)

			// Check for bug fixes (now including conventional commit type)
			if conventionalCommit != nil && conventionalCommit.Type == "fix" || 
			   a.isBugFix(commit.Commit.Message) {
				feature.Bugs = append(feature.Bugs, models.Bug{
					Description:   commit.Commit.Message,
					FixedAt:      commit.Commit.Author.When,
					CommitHash:   commit.Commit.Hash.String(),
					AuthorEmail:  commit.Commit.Author.Email,
					AffectedFiles: commit.Files,
				})
			}
		}
	}
}

func (a *Analyzer) matchesFeature(file string, patterns []*regexp.Regexp) bool {
	normalizedPath := filepath.ToSlash(file)
	for _, pattern := range patterns {
		if pattern.MatchString(normalizedPath) {
			return true
		}
	}
	return false
}

func (a *Analyzer) isBugFix(message string) bool {
	message = strings.ToLower(message)
	bugKeywords := []string{"fix", "bug", "issue", "resolve", "patch"}
	
	for _, keyword := range bugKeywords {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}

func (a *Analyzer) getAuthorCommitCounts(commits []git.CommitInfo) map[string]int {
	counts := make(map[string]int)
	for _, commit := range commits {
		counts[commit.Commit.Author.Email]++  // Using email instead of name
	}
	return counts
}

func parseConventionalCommit(message string) *ConventionalCommit {
	// Matches: <type>[optional scope][!]: <description>
	conventionalPattern := regexp.MustCompile(`^(?P<type>\w+)(?:\((?P<scope>[\w-]+)\))?(?P<breaking>!)?:\s*(?P<description>.+)`)
	
	lines := strings.Split(message, "\n")
	if len(lines) == 0 {
		return nil
	}

	matches := conventionalPattern.FindStringSubmatch(lines[0])
	if matches == nil {
		return nil
	}

	// Get named capture groups
	typeIdx := conventionalPattern.SubexpIndex("type")
	scopeIdx := conventionalPattern.SubexpIndex("scope")
	breakingIdx := conventionalPattern.SubexpIndex("breaking")
	descIdx := conventionalPattern.SubexpIndex("description")

	commit := &ConventionalCommit{
		Type:        matches[typeIdx],
		Description: matches[descIdx],
		Breaking:    matches[breakingIdx] != "",
	}

	if scopeIdx >= 0 && scopeIdx < len(matches) {
		commit.Scope = matches[scopeIdx]
	}

	// Check for body and BREAKING CHANGE footer
	if len(lines) > 1 {
		var bodyLines []string
		for _, line := range lines[1:] {
			if strings.HasPrefix(line, "BREAKING CHANGE:") {
				commit.Breaking = true
				continue
			}
			bodyLines = append(bodyLines, line)
		}
		commit.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	}

	return commit
} 