package mapper

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestMapperMap(t *testing.T) {
	testCases := []struct {
		name     string
		image    string
		repos    []Repo
		expected *Mapping
	}{
		{
			name:  "simple basename match",
			image: "nginx",
			repos: []Repo{
				{
					Name:        "nginx",
					CatalogTier: "APPLICATION",
					Aliases:     []string{},
				},
			},
			expected: &Mapping{
				Image:   "nginx",
				Results: []string{"nginx"},
			},
		},
		{
			name:  "no matches",
			image: "nonexistent",
			repos: []Repo{
				{
					Name:        "nginx",
					CatalogTier: "APPLICATION",
					Aliases:     []string{},
				},
			},
			expected: &Mapping{
				Image:   "nonexistent",
				Results: []string{},
			},
		},
		{
			name:  "multiple matches",
			image: "nginx",
			repos: []Repo{
				{
					Name:        "nginx",
					CatalogTier: "APPLICATION",
					Aliases:     []string{},
				},
				{
					Name:        "nginx-custom",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"nginx"},
				},
			},
			expected: &Mapping{
				Image:   "nginx",
				Results: []string{"nginx", "nginx-custom"},
			},
		},
		{
			name:  "tier filtering",
			image: "nginx",
			repos: []Repo{
				{
					Name:        "nginx",
					CatalogTier: "APPLICATION",
					Aliases:     []string{},
				},
				{
					Name:        "nginx-dev",
					CatalogTier: "FIPS",
					Aliases:     []string{"nginx"},
				},
			},
			expected: &Mapping{
				Image:   "nginx",
				Results: []string{"nginx"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &Mapper{
				repos:       tc.repos,
				ignoreTiers: []string{"fips"},
			}

			result, err := m.Map(tc.image)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Sort results for consistent comparison
			opts := cmpopts.SortSlices(func(a, b string) bool {
				return strings.Compare(a, b) < 0
			})

			if diff := cmp.Diff(tc.expected, result, opts); diff != "" {
				t.Errorf("mapping mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMapperMapInvalidImage(t *testing.T) {
	m := &Mapper{
		repos: []Repo{},
	}

	_, err := m.Map("invalid::image")
	if err == nil {
		t.Errorf("expected error for invalid image reference")
	}
}

func TestMapperMapAll(t *testing.T) {
	repos := []Repo{
		{
			Name:        "nginx",
			CatalogTier: "APPLICATION",
			Aliases:     []string{},
		},
		{
			Name:        "redis",
			CatalogTier: "APPLICATION",
			Aliases:     []string{},
		},
	}

	m := &Mapper{
		repos:       repos,
		ignoreTiers: []string{},
	}

	images := []string{"nginx", "redis", "postgres"}
	iterator := NewArgsIterator(images)

	results, err := m.MapAll(iterator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []*Mapping{
		{
			Image:   "nginx",
			Results: []string{"nginx"},
		},
		{
			Image:   "redis",
			Results: []string{"redis"},
		},
		{
			Image:   "postgres",
			Results: []string{},
		},
	}

	// Sort results for consistent comparison
	opts := cmpopts.SortSlices(func(a, b string) bool {
		return strings.Compare(a, b) < 0
	})

	if diff := cmp.Diff(expected, results, opts); diff != "" {
		t.Errorf("mapping results mismatch (-want +got):\n%s", diff)
	}
}

func TestMapperMapAllDuplicates(t *testing.T) {
	repos := []Repo{
		{
			Name:        "nginx",
			CatalogTier: "APPLICATION",
			Aliases:     []string{},
		},
	}

	m := &Mapper{
		repos:       repos,
		ignoreTiers: []string{},
	}

	// Include duplicates in the input
	images := []string{"nginx", "nginx", "redis"}
	iterator := NewArgsIterator(images)

	results, err := m.MapAll(iterator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have unique results
	expected := []*Mapping{
		{
			Image:   "nginx",
			Results: []string{"nginx"},
		},
		{
			Image:   "redis",
			Results: []string{},
		},
	}

	if len(results) != len(expected) {
		t.Errorf("expected %d results, got %d", len(expected), len(results))
	}

	// Sort results for consistent comparison
	opts := cmpopts.SortSlices(func(a, b string) bool {
		return strings.Compare(a, b) < 0
	})

	if diff := cmp.Diff(expected, results, opts); diff != "" {
		t.Errorf("mapping results mismatch (-want +got):\n%s", diff)
	}
}

func TestMapperMapAllIteratorError(t *testing.T) {
	m := &Mapper{
		repos:       []Repo{},
		ignoreTiers: []string{},
	}

	expectedErr := errors.New("iterator error")
	iterator := &errorIterator{err: expectedErr}

	_, err := m.MapAll(iterator)
	if err == nil {
		t.Error("expected error from iterator")
	}
}

func TestMapperMapAllMapError(t *testing.T) {
	m := &Mapper{
		repos:       []Repo{},
		ignoreTiers: []string{},
	}

	// Use an invalid image that will cause Map to fail
	images := []string{"invalid::image"}
	iterator := NewArgsIterator(images)

	_, err := m.MapAll(iterator)
	if err == nil {
		t.Error("expected error from Map")
	}
}

// errorIterator is a helper type for testing iterator errors
type errorIterator struct {
	err error
}

func (it *errorIterator) Next() (string, error) {
	return "", it.err
}
