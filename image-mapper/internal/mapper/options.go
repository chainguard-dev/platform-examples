package mapper

// Option configures a Mapper
type Option func(*options)

type options struct {
	ignoreFns []IgnoreFn
}

// WithIgnoreFns is a functional option that configures the IgnoreFns used by
// the mapper
func WithIgnoreFns(ignoreFns ...IgnoreFn) Option {
	return func(o *options) {
		o.ignoreFns = ignoreFns
	}
}
