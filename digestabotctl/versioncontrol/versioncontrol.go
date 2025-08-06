package versioncontrol

import (
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
)

type CommitOptions struct {
	Directory string
	Message   string
	Name      string
	Email     string
	When      time.Time
}

func Commit(opts CommitOptions) (string, error) {
	r, err := git.PlainOpen(opts.Directory)
	if err != nil {
		return "", err
	}

	w, err := r.Worktree()
	if err != nil {
		return "", err
	}

	if err := w.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return "", err
	}

	commitOptions := &git.CommitOptions{
		Author: &object.Signature{
			Name:  opts.Name,
			Email: opts.Email,
			When:  opts.When,
		},
	}

	hash, err := w.Commit(opts.Message, commitOptions)
	if err != nil {
		return "", err
	}

	commit, err := r.CommitObject(hash)
	if err != nil {
		return "", err
	}

	parent, err := commit.Parent(0)
	if err != nil {
		return "", err
	}
	patch, err := parent.Patch(commit)
	if err != nil {
		return "", err
	}

	return patch.String(), err
}
