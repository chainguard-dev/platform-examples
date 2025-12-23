package mapper

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

// Mapping describes an image and the Chainguard images it maps to
type Mapping struct {
	Image   string   `json:"image"`
	Results []string `json:"results,omitempty"`
}

// Mapper maps image references to images in our catalog
type Mapper interface {
	Map(image string) (*Mapping, error)
}

type mapper struct {
	repos       []Repo
	ignoreFns   []IgnoreFn
	includeTags []TagFilter
	repoName    string
}

// NewMapper creates a new mapper
func NewMapper(ctx context.Context, opts ...Option) (*mapper, error) {
	o := &options{
		repo: "cgr.dev/chainguard",
	}
	for _, opt := range opts {
		opt(o)
	}

	repoName, err := parseRepo(o.repo)
	if err != nil {
		return nil, fmt.Errorf("parsing repository: %w", err)
	}

	repos, err := listRepos(ctx, o.inactiveTags)
	if err != nil {
		return nil, fmt.Errorf("listing repos: %w", err)
	}

	m := &mapper{
		repos:       repos,
		ignoreFns:   o.ignoreFns,
		includeTags: o.includeTags,
		repoName:    repoName,
	}

	return m, nil
}

// MapAll returns mappings for all the images returned by the iterator
func (m *mapper) MapAll(it Iterator) ([]*Mapping, error) {
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
func (m *mapper) Map(image string) (*Mapping, error) {
	ref, err := name.NewTag(strings.Split(image, "@")[0])
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", image, err)
	}

	// Identify repositories in the Chainguard catalog that match the
	// provided image
	matches := map[string]Repo{}
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
		matches[cgrrepo.Name] = cgrrepo
	}

	// Format the matches into the results we'll include in the mappings
	results := []string{}
	for _, cgrrepo := range matches {
		// Append the repository name to the rest of the reference
		result := fmt.Sprintf("%s/%s", m.repoName, cgrrepo.Name)

		// Only match active tags unless we've fetched the full list of
		// tags
		tags := cgrrepo.ActiveTags
		if len(cgrrepo.Tags) > 0 {
			tags = flattenTags(cgrrepo.Tags)
		}
		tags = includeTags(tags, m.includeTags...)

		// Try and match the provided tag to one of the tags
		tag := MatchTag(tags, ref.TagStr())
		if tag != "" {
			result = fmt.Sprintf("%s:%s", result, tag)
		}
		results = append(results, result)
	}
	slices.Sort(results)

	return &Mapping{
		Image:   image,
		Results: results,
	}, nil
}

func (m *mapper) ignoreRepo(repo Repo) bool {
	for _, ignore := range m.ignoreFns {
		if !ignore(repo) {
			continue
		}
		return true
	}

	return false
}

// MapImage maps the provided image to its Chainguard equivalent. It returns the
// first result it finds.
func MapImage(m Mapper, img string) (name.Reference, error) {
	mapping, err := m.Map(img)
	if err != nil {
		return nil, fmt.Errorf("mapping image: %s: %w", img, err)
	}
	if len(mapping.Results) == 0 {
		return nil, fmt.Errorf("no results found")
	}
	result := mapping.Results[0]

	mapped, err := name.NewTag(result)
	if err != nil {
		return nil, fmt.Errorf("parsing mapped image: %w", err)
	}

	return mapped, nil
}
