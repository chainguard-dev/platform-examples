package mapper

import (
	"testing"
)

func TestMatchTag(t *testing.T) {
	activeTags := []string{
		"latest",
		"latest-dev",
		"3",
		"3.14",
		"3.14.2",
		"3.13",
		"3.13.6",
		"v3",
		"v3.14",
		"v3.14.2",
		"v3.13",
		"v3.13.6",
	}

	tests := []struct {
		name     string
		tag      string
		expected string
	}{
		{
			name:     "exact match patch version",
			tag:      "3.14.2",
			expected: "3.14.2",
		},
		{
			name:     "exact match minor version",
			tag:      "3.13",
			expected: "3.13",
		},
		{
			name:     "exact match major version",
			tag:      "3",
			expected: "3",
		},
		{
			name:     "nearest higher patch in same minor",
			tag:      "3.14.1",
			expected: "3.14.2",
		},
		{
			name:     "nearest higher patch in next minor",
			tag:      "3.12.5",
			expected: "3.13.6",
		},
		{
			name:     "nearest higher minor",
			tag:      "3.12",
			expected: "3.13",
		},
		{
			name:     "nearest higher major",
			tag:      "2",
			expected: "3",
		},
		{
			name:     "v-prefix exact match patch",
			tag:      "v3.14.2",
			expected: "v3.14.2",
		},
		{
			name:     "v-prefix exact match minor",
			tag:      "v3.13",
			expected: "v3.13",
		},
		{
			name:     "v-prefix exact match major",
			tag:      "v3",
			expected: "v3",
		},
		{
			name:     "v-prefix nearest higher patch",
			tag:      "v3.14.1",
			expected: "v3.14.2",
		},
		{
			name:     "v-prefix nearest higher minor",
			tag:      "v3.12",
			expected: "v3.13",
		},
		{
			name:     "v-prefix nearest higher major",
			tag:      "v2",
			expected: "v3",
		},
		{
			name:     "no match - tag too high",
			tag:      "4",
			expected: "",
		},
		{
			name:     "no match - v-prefix mismatch",
			tag:      "v4",
			expected: "",
		},
		{
			name:     "no match - minor too high",
			tag:      "3.15",
			expected: "",
		},
		{
			name:     "no match - patch too high",
			tag:      "3.14.3",
			expected: "",
		},
		{
			name:     "invalid tag",
			tag:      "invalid",
			expected: "",
		},
		{
			name:     "empty tag",
			tag:      "",
			expected: "",
		},
		{
			name:     "suffix exact match patch version",
			tag:      "3.14.2-alpine",
			expected: "3.14.2",
		},
		{
			name:     "suffix exact match minor version",
			tag:      "3.13-alpine",
			expected: "3.13",
		},
		{
			name:     "suffix exact match major version",
			tag:      "3-alpine",
			expected: "3",
		},
		{
			name:     "suffix v-prefix exact match patch",
			tag:      "v3.14.2-alpine",
			expected: "v3.14.2",
		},
		{
			name:     "suffix v-prefix exact match minor",
			tag:      "v3.13-alpine",
			expected: "v3.13",
		},
		{
			name:     "suffix v-prefix exact match major",
			tag:      "v3-alpine",
			expected: "v3",
		},
		{
			name:     "suffix nearest higher patch in same minor",
			tag:      "3.14.1-alpine",
			expected: "3.14.2",
		},
		{
			name:     "suffix nearest higher patch in next minor",
			tag:      "3.12.5-alpine",
			expected: "3.13.6",
		},
		{
			name:     "suffix nearest higher minor",
			tag:      "3.12-alpine",
			expected: "3.13",
		},
		{
			name:     "suffix nearest higher major",
			tag:      "2-alpine",
			expected: "3",
		},
		{
			name:     "suffix v-prefix nearest higher patch",
			tag:      "v3.14.1-alpine",
			expected: "v3.14.2",
		},
		{
			name:     "suffix v-prefix nearest higher minor",
			tag:      "v3.12-alpine",
			expected: "v3.13",
		},
		{
			name:     "suffix v-prefix nearest higher major",
			tag:      "v2-alpine",
			expected: "v3",
		},
		{
			name:     "suffix no match - tag too high",
			tag:      "4-alpine",
			expected: "",
		},
		{
			name:     "suffix no match - v-prefix mismatch",
			tag:      "v4-alpine",
			expected: "",
		},
		{
			name:     "suffix no match - minor too high",
			tag:      "3.15-alpine",
			expected: "",
		},
		{
			name:     "suffix no match - patch too high",
			tag:      "3.14.3-alpine",
			expected: "",
		},
		{
			name:     "suffix with multiple dashes",
			tag:      "3.14.2-alpine-slim",
			expected: "3.14.2",
		},
		{
			name:     "suffix v-prefix with multiple dashes",
			tag:      "v3.14.2-alpine-slim",
			expected: "v3.14.2",
		},
		{
			name:     "suffix debian",
			tag:      "3.14-debian",
			expected: "3.14",
		},
		{
			name:     "suffix slim",
			tag:      "3.14.2-slim",
			expected: "3.14.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchTag(activeTags, tt.tag)
			if result != tt.expected {
				t.Errorf("MatchTag(%q) = %q, expected %q", tt.tag, result, tt.expected)
			}
		})
	}
}
