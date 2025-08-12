package versioncontrol

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
)

type CommitOptions struct {
	Directory string
	Message   string
	Name      string
	Email     string
	When      time.Time
	Branch    string
	Token     string
}

type CheckoutResponse struct {
	Repo     *git.Repository
	Worktree *git.Worktree
}

func Checkout(opts CommitOptions) (CheckoutResponse, error) {
	r, err := git.PlainOpen(opts.Directory)
	if err != nil {
		return CheckoutResponse{}, err
	}

	w, err := r.Worktree()
	if err != nil {
		return CheckoutResponse{}, err
	}

	create := false
	branchRef := plumbing.NewBranchReferenceName(opts.Branch)
	_, err = r.Reference(branchRef, true)
	if err != nil && errors.Is(err, plumbing.ErrReferenceNotFound) {
		create = true
	} else if err != nil {
		return CheckoutResponse{}, err
	}

	if err := w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(opts.Branch),
		Create: create,
	}); err != nil {
		return CheckoutResponse{}, err
	}

	return CheckoutResponse{
		Repo:     r,
		Worktree: w,
	}, nil

}

func CommitAndPush(r *git.Repository, w *git.Worktree, opts CommitOptions) (string, error) {
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

	if err := r.Push(&git.PushOptions{
		RemoteName: "origin",
		Force:      true,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", opts.Branch, opts.Branch)),
		},
		// TODO: abstract this for other platforms
		Auth: &http.BasicAuth{
			Username: "x-access-token",
			Password: opts.Token,
		},
	}); err != nil {
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
