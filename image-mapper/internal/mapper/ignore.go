package mapper

import (
	"slices"
	"strings"
)

// IgnoreFn configures a mapper to ignore repositories
type IgnoreFn func(Repo) bool

// IgnoreTiers ignores repos that are in the provided tiers
func IgnoreTiers(tiers []string) IgnoreFn {
	var ignoreTiers []string
	for _, tier := range tiers {
		ignoreTiers = append(ignoreTiers, strings.ToLower(tier))
	}
	return func(repo Repo) bool {
		return slices.Contains(ignoreTiers, strings.ToLower(repo.CatalogTier))
	}
}

// IgnoreIamguarded ignores iamguarded repos
func IgnoreIamguarded() IgnoreFn {
	return func(repo Repo) bool {
		return strings.HasSuffix(repo.Name, "iamguarded") || strings.HasSuffix(repo.Name, "iamguarded-fips")
	}
}
