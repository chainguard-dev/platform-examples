package mapper

import (
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
)

func TestMatch(t *testing.T) {
	testCases := []struct {
		name     string
		refStr   string
		repo     *Repo
		expected bool
	}{
		{
			name:   "basename exact match",
			refStr: "nginx",
			repo: &Repo{
				Name: "nginx",
			},
			expected: true,
		},
		{
			name:   "basename with registry",
			refStr: "gcr.io/project/prometheus",
			repo: &Repo{
				Name: "prometheus",
			},
			expected: true,
		},
		{
			name:   "basename nested path",
			refStr: "bitnami/nginx",
			repo: &Repo{
				Name: "nginx",
			},
			expected: true,
		},
		{
			name:   "basename no match different name",
			refStr: "nginx",
			repo: &Repo{
				Name: "apache",
			},
			expected: false,
		},
		{
			name:   "basename no match partial name",
			refStr: "nginx-proxy",
			repo: &Repo{
				Name: "nginx",
			},
			expected: false,
		},
		{
			name:   "basename match iamguarded",
			refStr: "nginx",
			repo: &Repo{
				Name: "nginx-iamguarded",
			},
			expected: true,
		},
		{
			name:   "basename match iamguarded fips",
			refStr: "nginx",
			repo: &Repo{
				Name: "nginx-iamguarded-fips",
			},
			expected: true,
		},
		{
			name:   "basename match iamguarded registry path",
			refStr: "gcr.io/project/nginx",
			repo: &Repo{
				Name: "nginx-iamguarded",
			},
			expected: true,
		},
		{
			name:   "basename match iamguarded fips registry path",
			refStr: "gcr.io/project/nginx",
			repo: &Repo{
				Name: "nginx-iamguarded-fips",
			},
			expected: true,
		},
		{
			name:   "basename match fips",
			refStr: "nginx",
			repo: &Repo{
				Name: "nginx-fips",
			},
			expected: true,
		},
		{
			name:   "dashname",
			refStr: "ghcr.io/stakater/reloader",
			repo: &Repo{
				Name: "stakater-reloader",
			},
			expected: true,
		},
		{
			name:   "dashname three level conversion",
			refStr: "ghcr.io/foo/bar/baz",
			repo: &Repo{
				Name: "foo-bar-baz",
			},
			expected: true,
		},
		{
			name:   "dashname no match single name",
			refStr: "nginx",
			repo: &Repo{
				Name: "nginx-something",
			},
			expected: false,
		},
		{
			name:   "dashname no match different name",
			refStr: "stakater/reloader",
			repo: &Repo{
				Name: "different-reloader",
			},
			expected: false,
		},
		{
			name:   "dashname match fips",
			refStr: "stakater/reloader",
			repo: &Repo{
				Name: "stakater-reloader-fips",
			},
			expected: true,
		},
		{
			name:   "dashname match iamguarded",
			refStr: "stakater/reloader",
			repo: &Repo{
				Name: "stakater-reloader-iamguarded",
			},
			expected: true,
		},
		{
			name:   "dashname match iamguarded fips",
			refStr: "stakater/reloader",
			repo: &Repo{
				Name: "stakater-reloader-iamguarded-fips",
			},
			expected: true,
		},
		{
			name:   "dashname match iamguarded registry path",
			refStr: "registry.example.com/stakater/reloader",
			repo: &Repo{
				Name: "stakater-reloader-iamguarded",
			},
			expected: true,
		},
		{
			name:   "dashname match iamguarded fips registry path",
			refStr: "registry.example.com/stakater/reloader",
			repo: &Repo{
				Name: "stakater-reloader-iamguarded-fips",
			},
			expected: true,
		},
		{
			name:   "aliases exact repository match",
			refStr: "nginx",
			repo: &Repo{
				Name: "nginx-something",
				Aliases: []string{
					"nginx",
				},
			},
			expected: true,
		},
		{
			// Don't match basenames in aliases because they can be
			// very generic in some cases.
			name:   "aliases basename match",
			refStr: "custom.registry.com/cert-manager/controller",
			repo: &Repo{
				Name: "nginx-something",
				Aliases: []string{
					"fluxcd/controller",
				},
			},
			expected: false,
		},
		{
			name:   "aliases dash conversion match",
			refStr: "custom.registry.com/ingress-foobar",
			repo: &Repo{
				Name: "ingress-foobar-something",
				Aliases: []string{
					"ghcr.io/ingress/foobar",
				},
			},
			expected: true,
		},
		{
			name:   "aliases multiple first match",
			refStr: "nginx",
			repo: &Repo{
				Name: "nginx",
				Aliases: []string{
					"nginx",
					"gcr.io/project/nginx",
				},
			},
			expected: true,
		},
		{
			name:   "aliases multiple second match",
			refStr: "gcr.io/project/nginx",
			repo: &Repo{
				Name: "nginx",
				Aliases: []string{
					"apache",
					"gcr.io/project/nginx",
				},
			},
			expected: true,
		},
		{
			name:   "aliases no match different",
			refStr: "nginx",
			repo: &Repo{
				Name: "nginx-something",
				Aliases: []string{
					"apache",
					"gcr.io/project/httpd",
				},
			},
			expected: false,
		},
		{
			name:   "aliases complex dash conversion",
			refStr: "registry.example.com/org-suborg-project",
			repo: &Repo{
				Name: "project",
				Aliases: []string{
					"org/suborg/project",
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := name.ParseReference(tc.refStr)
			if err != nil {
				t.Fatalf("failed to parse reference %s: %v", tc.refStr, err)
			}

			result := Match(ref, *tc.repo)
			if result != tc.expected {
				t.Errorf("Match(%s, %+v) = %v, want %v", tc.refStr, tc.repo, result, tc.expected)
			}
		})
	}
}
