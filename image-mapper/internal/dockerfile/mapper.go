package dockerfile

import (
	"context"

	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/mapper"
)

// NewMapper returns a mapper.Mapper configured specifically for mapping images
// in Helm charts and values
func NewMapper(ctx context.Context, opts ...mapper.Option) (mapper.Mapper, error) {
	defaultOpts := []mapper.Option{
		mapper.WithIgnoreFns(
			// Iamguarded images are only designed to be
			// used with our Helm charts.
			mapper.IgnoreIamguarded(),
			// TODO: make it possible select only
			// FIPS images
			mapper.IgnoreTiers([]string{"FIPS"}),
		),
		// Use -dev tags because they're more likely to work out of the
		// box
		mapper.WithTagFilters(mapper.TagFilterPreferDev),
	}

	return mapper.NewMapper(ctx, append(defaultOpts, opts...)...)
}
