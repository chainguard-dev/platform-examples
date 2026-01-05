package helm

import (
	"context"
	"strings"

	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/mapper"
)

// NewMapper returns a mapper.Mapper configured specifically for mapping images
// in Helm charts and values
func NewMapper(ctx context.Context, opts ...mapper.Option) (mapper.Mapper, error) {
	defaultOpts := []mapper.Option{
		// Helm charts are designed to work with
		// specific versions. We include inactive tags
		// here so we can match to the closest
		// version.
		mapper.WithInactiveTags(true),
		mapper.WithIgnoreFns(
			// Iamguarded images are designed to be
			// used with our Helm charts.
			mapper.IgnoreIamguarded(),
			// TODO: make it possible select only
			// FIPS images
			mapper.IgnoreTiers([]string{"FIPS"}),
		),
		// Our non-dev tags *should* be able to be
		// dropped into upstream helm
		// charts, so let's prefer them by ensuring we
		// don't match to -dev tags.
		mapper.WithIncludeTags(
			func(tag string) bool {
				return !strings.HasSuffix(tag, "-dev")
			},
		),
	}

	return mapper.NewMapper(ctx, append(defaultOpts, opts...)...)
}
