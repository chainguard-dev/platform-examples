package mapper

// Option configures a Mapper
type Option func(*options)

type options struct {
	ignoreFns    []IgnoreFn
	repo         string
	inactiveTags bool
	tagFilters   []TagFilter
}

// WithIgnoreFns is a functional option that configures the IgnoreFns used by
// the mapper
func WithIgnoreFns(ignoreFns ...IgnoreFn) Option {
	return func(o *options) {
		o.ignoreFns = ignoreFns
	}
}

// WithRepository is a functional option that configures the repository prefix
// of the returned results
func WithRepository(repo string) Option {
	return func(o *options) {
		o.repo = repo
	}
}

// WithTagFilters is a functional option that configures tag filters to apply to
// matches
func WithTagFilters(tagFilters ...TagFilter) Option {
	return func(o *options) {
		o.tagFilters = tagFilters
	}
}

// WithInactiveTags is a functional option that configures the mapper to include
// inactive tags in its matching
func WithInactiveTags(inactiveTags bool) Option {
	return func(o *options) {
		o.inactiveTags = inactiveTags
	}
}
