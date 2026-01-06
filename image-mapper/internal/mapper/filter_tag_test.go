package mapper

import (
	"reflect"
	"testing"
)

func TestTagFilterExcludeDev(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected []string
	}{
		{
			name:     "empty tags",
			tags:     []string{},
			expected: nil,
		},
		{
			name:     "no dev tags",
			tags:     []string{"v1.0.0", "v2.0.0", "latest"},
			expected: []string{"v1.0.0", "v2.0.0", "latest"},
		},
		{
			name:     "all dev tags",
			tags:     []string{"v1.0.0-dev", "v2.0.0-dev", "latest-dev"},
			expected: nil,
		},
		{
			name:     "mixed dev and non-dev tags",
			tags:     []string{"v1.0.0", "v1.0.0-dev", "v2.0.0", "v2.0.0-dev", "latest"},
			expected: []string{"v1.0.0", "v2.0.0", "latest"},
		},
		{
			name:     "tag containing dev but not ending with -dev",
			tags:     []string{"development", "v1.0.0-dev", "devops"},
			expected: []string{"development", "devops"},
		},
		{
			name:     "single dev tag",
			tags:     []string{"v1.0.0-dev"},
			expected: nil,
		},
		{
			name:     "single non-dev tag",
			tags:     []string{"v1.0.0"},
			expected: []string{"v1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TagFilterExcludeDev(tt.tags)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("TagFilterExcludeDev() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTagFilterIncludeDev(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected []string
	}{
		{
			name:     "empty tags",
			tags:     []string{},
			expected: nil,
		},
		{
			name:     "no dev tags",
			tags:     []string{"v1.0.0", "v2.0.0", "latest"},
			expected: nil,
		},
		{
			name:     "all dev tags",
			tags:     []string{"v1.0.0-dev", "v2.0.0-dev", "latest-dev"},
			expected: []string{"v1.0.0-dev", "v2.0.0-dev", "latest-dev"},
		},
		{
			name:     "mixed dev and non-dev tags",
			tags:     []string{"v1.0.0", "v1.0.0-dev", "v2.0.0", "v2.0.0-dev", "latest"},
			expected: []string{"v1.0.0-dev", "v2.0.0-dev"},
		},
		{
			name:     "tag containing dev but not ending with -dev",
			tags:     []string{"development", "v1.0.0-dev", "devops"},
			expected: []string{"v1.0.0-dev"},
		},
		{
			name:     "single dev tag",
			tags:     []string{"v1.0.0-dev"},
			expected: []string{"v1.0.0-dev"},
		},
		{
			name:     "single non-dev tag",
			tags:     []string{"v1.0.0"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TagFilterIncludeDev(tt.tags)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("TagFilterIncludeDev() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTagFilterPreferDev(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected []string
	}{
		{
			name:     "empty tags",
			tags:     []string{},
			expected: []string{},
		},
		{
			name:     "no dev tags returns all tags",
			tags:     []string{"v1.0.0", "v2.0.0", "latest"},
			expected: []string{"v1.0.0", "v2.0.0", "latest"},
		},
		{
			name:     "all dev tags returns all tags",
			tags:     []string{"v1.0.0-dev", "v2.0.0-dev", "latest-dev"},
			expected: []string{"v1.0.0-dev", "v2.0.0-dev", "latest-dev"},
		},
		{
			name:     "mixed dev and non-dev tags returns only dev tags",
			tags:     []string{"v1.0.0", "v1.0.0-dev", "v2.0.0", "v2.0.0-dev", "latest"},
			expected: []string{"v1.0.0-dev", "v2.0.0-dev"},
		},
		{
			name:     "tag containing dev but not ending with -dev returns all",
			tags:     []string{"development", "devops", "production"},
			expected: []string{"development", "devops", "production"},
		},
		{
			name:     "single dev tag returns only dev tag",
			tags:     []string{"v1.0.0-dev"},
			expected: []string{"v1.0.0-dev"},
		},
		{
			name:     "single non-dev tag returns that tag",
			tags:     []string{"v1.0.0"},
			expected: []string{"v1.0.0"},
		},
		{
			name:     "one dev tag among many non-dev returns only dev",
			tags:     []string{"v1.0.0", "v2.0.0", "v3.0.0-dev", "v4.0.0", "latest"},
			expected: []string{"v3.0.0-dev"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TagFilterPreferDev(tt.tags)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("TagFilterPreferDev() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
