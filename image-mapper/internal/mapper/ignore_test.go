package mapper

import (
	"testing"
)

func TestIgnoreTiers(t *testing.T) {
	tests := []struct {
		name       string
		tiers      []string
		repo       Repo
		wantIgnore bool
	}{
		{
			name:  "exact match - FIPS",
			tiers: []string{"FIPS"},
			repo: Repo{
				Name:        "test-repo",
				CatalogTier: "FIPS",
			},
			wantIgnore: true,
		},
		{
			name:  "case insensitive match - lowercase tier input",
			tiers: []string{"fips"},
			repo: Repo{
				Name:        "test-repo",
				CatalogTier: "FIPS",
			},
			wantIgnore: true,
		},
		{
			name:  "case insensitive match - lowercase repo tier",
			tiers: []string{"FIPS"},
			repo: Repo{
				Name:        "test-repo",
				CatalogTier: "fips",
			},
			wantIgnore: true,
		},
		{
			name:  "case insensitive match - mixed case",
			tiers: []string{"FiPs"},
			repo: Repo{
				Name:        "test-repo",
				CatalogTier: "fIpS",
			},
			wantIgnore: true,
		},
		{
			name:  "multiple tiers - first matches",
			tiers: []string{"FIPS", "BASE", "APPLICATION"},
			repo: Repo{
				Name:        "test-repo",
				CatalogTier: "FIPS",
			},
			wantIgnore: true,
		},
		{
			name:  "multiple tiers - middle matches",
			tiers: []string{"FIPS", "BASE", "APPLICATION"},
			repo: Repo{
				Name:        "test-repo",
				CatalogTier: "BASE",
			},
			wantIgnore: true,
		},
		{
			name:  "multiple tiers - last matches",
			tiers: []string{"FIPS", "BASE", "APPLICATION"},
			repo: Repo{
				Name:        "test-repo",
				CatalogTier: "APPLICATION",
			},
			wantIgnore: true,
		},
		{
			name:  "no match",
			tiers: []string{"FIPS"},
			repo: Repo{
				Name:        "test-repo",
				CatalogTier: "BASE",
			},
			wantIgnore: false,
		},
		{
			name:  "multiple tiers - no match",
			tiers: []string{"FIPS", "BASE", "APPLICATION"},
			repo: Repo{
				Name:        "test-repo",
				CatalogTier: "AI",
			},
			wantIgnore: false,
		},
		{
			name:  "empty tier list",
			tiers: []string{},
			repo: Repo{
				Name:        "test-repo",
				CatalogTier: "FIPS",
			},
			wantIgnore: false,
		},
		{
			name:  "empty catalog tier",
			tiers: []string{"FIPS"},
			repo: Repo{
				Name:        "test-repo",
				CatalogTier: "",
			},
			wantIgnore: false,
		},
		{
			name:  "empty string in tiers list matches empty catalog tier",
			tiers: []string{""},
			repo: Repo{
				Name:        "test-repo",
				CatalogTier: "",
			},
			wantIgnore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ignoreFn := IgnoreTiers(tt.tiers)
			got := ignoreFn(tt.repo)
			if got != tt.wantIgnore {
				t.Errorf("IgnoreTiers() = %v, want %v", got, tt.wantIgnore)
			}
		})
	}
}

func TestIgnoreIamguarded(t *testing.T) {
	tests := []struct {
		name       string
		repo       Repo
		wantIgnore bool
	}{
		{
			name: "repo ending with iamguarded",
			repo: Repo{
				Name: "test-repo-iamguarded",
			},
			wantIgnore: true,
		},
		{
			name: "repo ending with iamguarded-fips",
			repo: Repo{
				Name: "test-repo-iamguarded-fips",
			},
			wantIgnore: true,
		},
		{
			name: "repo with just iamguarded",
			repo: Repo{
				Name: "iamguarded",
			},
			wantIgnore: true,
		},
		{
			name: "repo with just iamguarded-fips",
			repo: Repo{
				Name: "iamguarded-fips",
			},
			wantIgnore: true,
		},
		{
			name: "repo not ending with iamguarded",
			repo: Repo{
				Name: "test-repo",
			},
			wantIgnore: false,
		},
		{
			name: "repo containing iamguarded but not at end",
			repo: Repo{
				Name: "iamguarded-test-repo",
			},
			wantIgnore: false,
		},
		{
			name: "repo containing iamguarded-fips but not at end",
			repo: Repo{
				Name: "iamguarded-fips-test",
			},
			wantIgnore: false,
		},
		{
			name: "empty repo name",
			repo: Repo{
				Name: "",
			},
			wantIgnore: false,
		},
		{
			name: "case sensitive - uppercase IAMGUARDED",
			repo: Repo{
				Name: "test-repo-IAMGUARDED",
			},
			wantIgnore: false,
		},
		{
			name: "case sensitive - uppercase IAMGUARDED-FIPS",
			repo: Repo{
				Name: "test-repo-IAMGUARDED-FIPS",
			},
			wantIgnore: false,
		},
		{
			name: "partial match - iamguarde (missing d)",
			repo: Repo{
				Name: "test-repo-iamguarde",
			},
			wantIgnore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ignoreFn := IgnoreIamguarded()
			got := ignoreFn(tt.repo)
			if got != tt.wantIgnore {
				t.Errorf("IgnoreIamguarded() = %v, want %v", got, tt.wantIgnore)
			}
		})
	}
}
