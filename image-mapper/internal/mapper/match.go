package mapper

import (
	"fmt"
	"path"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

// Match returns true if the container image described by the reference
// matches the provided Chainguard repostory
func Match(ref name.Reference, repo Repo) bool {
	for _, fn := range matchFns {
		if !fn(ref, repo) {
			continue
		}

		return true
	}

	return false
}

// MatchFn checks whether a given reference corresponds to a Chainguard repo
type MatchFn func(ref name.Reference, repo Repo) bool

var matchFns = []MatchFn{
	matchBasename,
	matchDashname,
	matchIamguarded,
	matchAliases,
}

// matchBasename matches Chainguard images that match the basename of the
// upstream repository. For instance, ghcr.io/foo/bar/nginx -> nginx.
func matchBasename(ref name.Reference, repo Repo) bool {
	basename := path.Base(ref.Context().String())
	if basename == repo.Name {
		return true
	}
	if fmt.Sprintf("%s-fips", basename) == repo.Name {
		return true
	}

	return false
}

// matchDashname matches Chainguard images that match the name of the upstream
// repository, joined by dashes. For instance, ghcr.io/stakater/reloader ->
// stakater-reloader.
func matchDashname(ref name.Reference, repo Repo) bool {
	dashname := strings.ReplaceAll(ref.Context().RepositoryStr(), "/", "-")
	if dashname == repo.Name {
		return true
	}
	if fmt.Sprintf("%s-fips", dashname) == repo.Name {
		return true
	}

	return false
}

// matchIamguarded matches Chainguard images that match the name of the upstream
// repository, with 'iamguarded' appended. This identifies iamguarded
// equivalents for upstream images.
func matchIamguarded(ref name.Reference, repo Repo) bool {
	withIamguarded := fmt.Sprintf("%s-iamguarded", ref.Context().String())

	iamguardedRef, err := name.ParseReference(withIamguarded)
	if err != nil {
		return false
	}

	return matchBasename(iamguardedRef, repo) || matchDashname(iamguardedRef, repo)
}

// matchAliases uses the Chainguard repository's aliases to match against the
// upstream reference
func matchAliases(ref name.Reference, repo Repo) bool {
	urepo := ref.Context().String()
	urepoStr := ref.Context().RepositoryStr()

	for _, alias := range repo.Aliases {
		aref, err := name.ParseReference(alias)
		if err != nil {
			continue
		}
		arepo := aref.Context().String()
		arepoStr := aref.Context().RepositoryStr()
		arepoDashStr := strings.ReplaceAll(arepoStr, "/", "-")

		// Match if the full repository (ghcr.io/foo/bar) matches the
		// alias.
		if urepo == arepo {
			return true
		}

		// Match if the repository name (foo/bar) matches the repository
		// name of the alias. This might be the case if the customer is
		// mirroring an upstream image into another registry.
		if urepoStr == arepoStr {
			return true
		}

		// Match if upstream repository name (foo-bar) matches the
		// repository name of the alias (foo/bar) with dashes instead of
		// /. This could happen if a customer is copying an image to a
		// mirror and flattening the name.
		if urepoStr == arepoDashStr {
			return true
		}

	}

	return false
}
