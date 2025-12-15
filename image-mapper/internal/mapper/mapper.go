package mapper

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/go-containerregistry/pkg/name"
)

// Mapping describes an image and the Chainguard images it maps to
type Mapping struct {
	Image   string   `json:"image"`
	Results []string `json:"results,omitempty"`
}

// Mapper maps image references to images in our catalog
type Mapper struct {
	repos     []Repo
	ignoreFns []IgnoreFn
}

// NewMapper creates a new mapper
func NewMapper(ctx context.Context, opts ...Option) (*Mapper, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	repos, err := listRepos(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing repos: %w", err)
	}

	m := &Mapper{
		repos:     repos,
		ignoreFns: o.ignoreFns,
	}

	return m, nil
}

// MapAll returns mappings for all the images returned by the iterator
func (m *Mapper) MapAll(it Iterator) ([]*Mapping, error) {
	mapped := make(map[string]struct{})
	mappings := []*Mapping{}
	for {
		image, err := it.Next()
		if err == ErrIteratorDone {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iterating over images: %w", err)
		}

		if _, ok := mapped[image]; ok {
			continue
		}

		mapping, err := m.Map(image)
		if err != nil {
			return nil, fmt.Errorf("mapping image %s: %w", image, err)
		}

		mappings = append(mappings, mapping)
		mapped[image] = struct{}{}
	}

	return mappings, nil
}

// Map an upstream image to the corresponding images in chainguard-private
func (m *Mapper) Map(image string) (*Mapping, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", image, err)
	}

	matches := map[string]struct{}{}
	for _, cgrrepo := range m.repos {
		// There are some images that may appear in the results but are
		// not accessible in the catalog. We can exclude them by
		// ignoring repos without a catalog tier.
		if cgrrepo.CatalogTier == "" {
			continue
		}

		if m.ignoreRepo(cgrrepo) {
			continue
		}

		if !Match(ref, cgrrepo) {
			continue
		}
		matches[cgrrepo.Name] = struct{}{}
	}

	results := []string{}
	for match := range matches {
		results = append(results, match)
	}
	slices.Sort(results)

	return &Mapping{
		Image:   image,
		Results: results,
	}, nil
}

func (m *Mapper) ignoreRepo(repo Repo) bool {
	for _, ignore := range m.ignoreFns {
		if !ignore(repo) {
			continue
		}
		return true
	}

	return false
}
