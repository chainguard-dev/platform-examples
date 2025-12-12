package mapper

import "strings"

// Option configures a Mapper
type Option func(*options)

type options struct {
	ignoreTiers []string
}

// WithoutTiers is a functional option that configures a Mapper to ignore
// Chainguard images of specific tiers
func WithoutTiers(tiers []string) Option {
	var ignoreTiers []string
	for _, tier := range tiers {
		ignoreTiers = append(ignoreTiers, strings.ToLower(tier))
	}
	return func(o *options) {
		o.ignoreTiers = ignoreTiers
	}
}
