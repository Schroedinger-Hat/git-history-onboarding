package git

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/go-git/go-billy/v5/memfs"
)

type Repository struct {
	repo *git.Repository
}

type CommitInfo struct {
	Commit *object.Commit
	Files  []string
}

// Clone clones a git repository into memory
func Clone(ctx context.Context, url string) (*Repository, error) {
	// Create memory storage and filesystem
	storage := memory.NewStorage()
	fs := memfs.New()

	// Clone the repository
	repo, err := git.CloneContext(ctx, storage, fs, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return nil, err
	}

	return &Repository{repo: repo}, nil
}

// GetCommitHistory returns all commits from the main branch
func (r *Repository) GetCommitHistory() ([]CommitInfo, error) {
	// Get reference to HEAD
	ref, err := r.repo.Head()
	if err != nil {
		return nil, err
	}

	// Get commit history
	cIter, err := r.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}

	var commits []CommitInfo
	err = cIter.ForEach(func(c *object.Commit) error {
		// Get files changed in this commit
		files, err := getChangedFiles(c)
		if err != nil {
			return err
		}

		commits = append(commits, CommitInfo{
			Commit: c,
			Files:  files,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return commits, nil
}

func getChangedFiles(commit *object.Commit) ([]string, error) {
	files := make([]string, 0)
	
	// Get parent commit to compare changes
	if commit.NumParents() > 0 {
		parent, err := commit.Parents().Next()
		if err != nil {
			return nil, err
		}

		// Get changes between parent and current commit
		patch, err := commit.Patch(parent)
		if err != nil {
			return nil, err
		}

		// Extract file names from patches
		for _, filePatch := range patch.FilePatches() {
			from, to := filePatch.Files()
			if from != nil {
				files = append(files, from.Path())
			}
			if to != nil && (from == nil || to.Path() != from.Path()) {
				files = append(files, to.Path())
			}
		}
	} else {
		// For first commit, get all files
		tree, err := commit.Tree()
		if err != nil {
			return nil, err
		}

		tree.Files().ForEach(func(f *object.File) error {
			files = append(files, f.Name)
			return nil
		})
	}

	return files, nil
} 