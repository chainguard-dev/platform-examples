package mapper

import (
	"reflect"
	"strings"
	"testing"
)

func TestIncludeTags(t *testing.T) {
	tags := []string{"latest", "v1.0.0", "v2.0.0", "dev", "prod", "staging"}

	tests := []struct {
		name     string
		tags     []string
		filters  []TagFilter
		expected []string
	}{
		{
			name:     "no filters returns all tags",
			tags:     tags,
			filters:  nil,
			expected: tags,
		},
		{
			name:     "empty filters returns all tags",
			tags:     tags,
			filters:  []TagFilter{},
			expected: tags,
		},
		{
			name: "single filter includes matching tags",
			tags: tags,
			filters: []TagFilter{
				func(tag string) bool { return strings.HasPrefix(tag, "v") },
			},
			expected: []string{"v1.0.0", "v2.0.0"},
		},
		{
			name: "multiple filters use OR logic",
			tags: tags,
			filters: []TagFilter{
				func(tag string) bool { return strings.HasPrefix(tag, "v") },
				func(tag string) bool { return tag == "dev" },
			},
			expected: []string{"v1.0.0", "v2.0.0", "dev"},
		},
		{
			name: "no tags match filters",
			tags: tags,
			filters: []TagFilter{
				func(tag string) bool { return strings.HasPrefix(tag, "nonexistent") },
			},
			expected: nil,
		},
		{
			name: "empty tags slice",
			tags: []string{},
			filters: []TagFilter{
				func(tag string) bool { return true },
			},
			expected: nil,
		},
		{
			name: "filter returns true for all",
			tags: tags,
			filters: []TagFilter{
				func(tag string) bool { return true },
			},
			expected: tags,
		},
		{
			name: "filter returns false for all",
			tags: tags,
			filters: []TagFilter{
				func(tag string) bool { return false },
			},
			expected: nil,
		},
		{
			name: "three filters with different matches",
			tags: []string{"alpha", "beta", "gamma", "delta"},
			filters: []TagFilter{
				func(tag string) bool { return tag == "alpha" },
				func(tag string) bool { return tag == "gamma" },
				func(tag string) bool { return strings.Contains(tag, "et") },
			},
			expected: []string{"alpha", "beta", "gamma"},
		},
		{
			name: "filter by tag length",
			tags: []string{"a", "ab", "abc", "abcd"},
			filters: []TagFilter{
				func(tag string) bool { return len(tag) > 2 },
			},
			expected: []string{"abc", "abcd"},
		},
		{
			name: "multiple filters where none match",
			tags: tags,
			filters: []TagFilter{
				func(tag string) bool { return strings.HasPrefix(tag, "x") },
				func(tag string) bool { return strings.HasPrefix(tag, "y") },
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := includeTags(tt.tags, tt.filters...)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("includeTags() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
