package versioncontrol

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type CommitOptions struct {
	Directory string
	Message   string
	Name      string
	Email     string
	When      time.Time
	Branch    string
	Token     string
	Signer    git.Signer
}

type CheckoutResponse struct {
	Repo     *git.Repository
	Worktree *git.Worktree
}

func Checkout(opts CommitOptions) (CheckoutResponse, error) {
	r, err := git.PlainOpen(opts.Directory)
	if err != nil {
		return CheckoutResponse{}, fmt.Errorf("git open: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return CheckoutResponse{}, fmt.Errorf("git worktree: %w", err)
	}

	create := false
	branchRef := plumbing.NewBranchReferenceName(opts.Branch)
	_, err = r.Reference(branchRef, true)
	if err != nil && errors.Is(err, plumbing.ErrReferenceNotFound) {
		create = true
	} else if err != nil {
		return CheckoutResponse{}, fmt.Errorf("git branch reference: %w", err)
	}

	if err := w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(opts.Branch),
		Create: create,
	}); err != nil {
		return CheckoutResponse{}, fmt.Errorf("git checkout: %w", err)
	}

	return CheckoutResponse{
		Repo:     r,
		Worktree: w,
	}, nil

}

func CommitAndPush(r *git.Repository, w *git.Worktree, opts CommitOptions) (string, error) {
	if err := w.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return "", fmt.Errorf("git add: %w", err)
	}

	commitOptions := &git.CommitOptions{
		All:    true,
		Signer: opts.Signer,
		Author: &object.Signature{
			Name:  opts.Name,
			Email: opts.Email,
			When:  opts.When,
		},
	}

	hash, err := w.Commit(fmt.Sprintf("%s\n", opts.Message), commitOptions)
	if err != nil {
		return "", fmt.Errorf("commit hash: %w", err)
	}

	commit, err := r.CommitObject(hash)
	if err != nil {
		return "", fmt.Errorf("git commit: %w", err)
	}

	if err := r.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", opts.Branch, opts.Branch)),
		},
		// TODO: abstract this for other platforms
		Auth: &http.BasicAuth{
			Username: "x-access-token",
			Password: opts.Token,
		},
	}); err != nil {
		return "", fmt.Errorf("git push: %w", err)
	}

	parent, err := commit.Parent(0)
	if err != nil {
		return "", fmt.Errorf("git parent: %w", err)
	}
	patch, err := parent.Patch(commit)
	if err != nil {
		return "", fmt.Errorf("commit patch: %w", err)
	}

	return patch.String(), err
}
