package mapper

import (
	"errors"
	"os"
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
				Results: []string{"cgr.dev/chainguard/nginx"},
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
				Results: []string{"cgr.dev/chainguard/nginx", "cgr.dev/chainguard/nginx-custom"},
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
				Results: []string{"cgr.dev/chainguard/nginx"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &mapper{
				repos:     tc.repos,
				repoName:  "cgr.dev/chainguard",
				ignoreFns: []IgnoreFn{IgnoreTiers([]string{"fips"})},
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
	m := &mapper{
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

	m := &mapper{
		repos:    repos,
		repoName: "cgr.dev/chainguard",
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
			Results: []string{"cgr.dev/chainguard/nginx"},
		},
		{
			Image:   "redis",
			Results: []string{"cgr.dev/chainguard/redis"},
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

	m := &mapper{
		repos:    repos,
		repoName: "cgr.dev/chainguard",
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
			Results: []string{"cgr.dev/chainguard/nginx"},
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
	m := &mapper{
		repos: []Repo{},
	}
	expectedErr := errors.New("iterator error")
	iterator := &errorIterator{err: expectedErr}

	_, err := m.MapAll(iterator)
	if err == nil {
		t.Error("expected error from iterator")
	}
}

func TestMapperMapAllMapError(t *testing.T) {
	m := &mapper{
		repos: []Repo{},
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

func TestMapperMapWithCustomIgnoreFn(t *testing.T) {
	testCases := []struct {
		name      string
		image     string
		repos     []Repo
		ignoreFns []IgnoreFn
		expected  *Mapping
	}{
		{
			name:  "ignore repos by name prefix",
			image: "nginx",
			repos: []Repo{
				{
					Name:        "nginx",
					CatalogTier: "APPLICATION",
					Aliases:     []string{},
				},
				{
					Name:        "test-nginx",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"nginx"},
				},
				{
					Name:        "prod-nginx",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"nginx"},
				},
			},
			ignoreFns: []IgnoreFn{
				func(repo Repo) bool {
					return strings.HasPrefix(repo.Name, "test-")
				},
			},
			expected: &Mapping{
				Image:   "nginx",
				Results: []string{"cgr.dev/chainguard/nginx", "cgr.dev/chainguard/prod-nginx"},
			},
		},
		{
			name:  "ignore repos containing specific string",
			image: "redis",
			repos: []Repo{
				{
					Name:        "redis",
					CatalogTier: "APPLICATION",
					Aliases:     []string{},
				},
				{
					Name:        "redis-dev",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"redis"},
				},
				{
					Name:        "redis-prod",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"redis"},
				},
			},
			ignoreFns: []IgnoreFn{
				func(repo Repo) bool {
					return strings.Contains(repo.Name, "-dev")
				},
			},
			expected: &Mapping{
				Image:   "redis",
				Results: []string{"cgr.dev/chainguard/redis", "cgr.dev/chainguard/redis-prod"},
			},
		},
		{
			name:  "multiple custom ignore functions",
			image: "postgres",
			repos: []Repo{
				{
					Name:        "postgres",
					CatalogTier: "APPLICATION",
					Aliases:     []string{},
				},
				{
					Name:        "test-postgres",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"postgres"},
				},
				{
					Name:        "postgres-dev",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"postgres"},
				},
				{
					Name:        "postgres-prod",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"postgres"},
				},
			},
			ignoreFns: []IgnoreFn{
				func(repo Repo) bool {
					return strings.HasPrefix(repo.Name, "test-")
				},
				func(repo Repo) bool {
					return strings.Contains(repo.Name, "-dev")
				},
			},
			expected: &Mapping{
				Image:   "postgres",
				Results: []string{"cgr.dev/chainguard/postgres", "cgr.dev/chainguard/postgres-prod"},
			},
		},
		{
			name:  "ignore repos by exact name match",
			image: "mysql",
			repos: []Repo{
				{
					Name:        "mysql",
					CatalogTier: "APPLICATION",
					Aliases:     []string{},
				},
				{
					Name:        "mysql-legacy",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"mysql"},
				},
				{
					Name:        "mysql-new",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"mysql"},
				},
			},
			ignoreFns: []IgnoreFn{
				func(repo Repo) bool {
					return repo.Name == "mysql-legacy"
				},
			},
			expected: &Mapping{
				Image:   "mysql",
				Results: []string{"cgr.dev/chainguard/mysql", "cgr.dev/chainguard/mysql-new"},
			},
		},
		{
			name:  "ignore all matching repos with custom function",
			image: "alpine",
			repos: []Repo{
				{
					Name:        "alpine-dev",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"alpine"},
				},
				{
					Name:        "alpine-test",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"alpine"},
				},
			},
			ignoreFns: []IgnoreFn{
				func(repo Repo) bool {
					return strings.HasPrefix(repo.Name, "alpine-")
				},
			},
			expected: &Mapping{
				Image:   "alpine",
				Results: []string{},
			},
		},
		{
			name:  "combine built-in and custom ignore functions",
			image: "node",
			repos: []Repo{
				{
					Name:        "node",
					CatalogTier: "APPLICATION",
					Aliases:     []string{},
				},
				{
					Name:        "node-fips",
					CatalogTier: "FIPS",
					Aliases:     []string{"node"},
				},
				{
					Name:        "experimental-node",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"node"},
				},
				{
					Name:        "node-staging",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"node"},
				},
			},
			ignoreFns: []IgnoreFn{
				IgnoreTiers([]string{"fips"}),
				func(repo Repo) bool {
					return strings.HasPrefix(repo.Name, "experimental-")
				},
			},
			expected: &Mapping{
				Image:   "node",
				Results: []string{"cgr.dev/chainguard/node", "cgr.dev/chainguard/node-staging"},
			},
		},
		{
			name:  "ignore repos by suffix",
			image: "python",
			repos: []Repo{
				{
					Name:        "python",
					CatalogTier: "APPLICATION",
					Aliases:     []string{},
				},
				{
					Name:        "python-slim",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"python"},
				},
				{
					Name:        "python-alpine",
					CatalogTier: "APPLICATION",
					Aliases:     []string{"python"},
				},
			},
			ignoreFns: []IgnoreFn{
				func(repo Repo) bool {
					return strings.HasSuffix(repo.Name, "-alpine")
				},
			},
			expected: &Mapping{
				Image:   "python",
				Results: []string{"cgr.dev/chainguard/python", "cgr.dev/chainguard/python-slim"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &mapper{
				repos:     tc.repos,
				repoName:  "cgr.dev/chainguard",
				ignoreFns: tc.ignoreFns,
			}

			result, err := m.Map(tc.image)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			opts := cmpopts.SortSlices(func(a, b string) bool {
				return strings.Compare(a, b) < 0
			})

			if diff := cmp.Diff(tc.expected, result, opts); diff != "" {
				t.Errorf("mapping mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMapperMapWithCustomIgnoreFnUsingAliases(t *testing.T) {
	repos := []Repo{
		{
			Name:        "web-server",
			CatalogTier: "APPLICATION",
			Aliases:     []string{"nginx", "httpd"},
		},
		{
			Name:        "cache-server",
			CatalogTier: "APPLICATION",
			Aliases:     []string{"redis", "memcached"},
		},
		{
			Name:        "db-server",
			CatalogTier: "APPLICATION",
			Aliases:     []string{"postgres", "mysql"},
		},
	}

	// Custom ignore function that checks aliases
	ignoreFn := func(repo Repo) bool {
		for _, alias := range repo.Aliases {
			if alias == "redis" || alias == "memcached" {
				return true
			}
		}
		return false
	}

	m := &mapper{
		repos:     repos,
		repoName:  "cgr.dev/chainguard",
		ignoreFns: []IgnoreFn{ignoreFn},
	}

	// Should match cache-server but it should be ignored
	result, err := m.Map("redis")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := &Mapping{
		Image:   "redis",
		Results: []string{},
	}

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("mapping mismatch (-want +got):\n%s", diff)
	}

	// Should match web-server and it should not be ignored
	result, err = m.Map("nginx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected = &Mapping{
		Image:   "nginx",
		Results: []string{"cgr.dev/chainguard/web-server"},
	}

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("mapping mismatch (-want +got):\n%s", diff)
	}
}

func TestMapperMapWithNoIgnoreFns(t *testing.T) {
	repos := []Repo{
		{
			Name:        "nginx",
			CatalogTier: "APPLICATION",
			Aliases:     []string{},
		},
		{
			Name:        "nginx-test",
			CatalogTier: "APPLICATION",
			Aliases:     []string{"nginx"},
		},
		{
			Name:        "nginx-dev",
			CatalogTier: "APPLICATION",
			Aliases:     []string{"nginx"},
		},
	}

	m := &mapper{
		repos:     repos,
		repoName:  "cgr.dev/chainguard",
		ignoreFns: []IgnoreFn{}, // No ignore functions
	}

	result, err := m.Map("nginx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should get all matching repos when no ignore functions are set
	expected := &Mapping{
		Image:   "nginx",
		Results: []string{"cgr.dev/chainguard/nginx", "cgr.dev/chainguard/nginx-dev", "cgr.dev/chainguard/nginx-test"},
	}

	opts := cmpopts.SortSlices(func(a, b string) bool {
		return strings.Compare(a, b) < 0
	})

	if diff := cmp.Diff(expected, result, opts); diff != "" {
		t.Errorf("mapping mismatch (-want +got):\n%s", diff)
	}
}

func TestMapImage(t *testing.T) {
	testCases := []struct {
		name          string
		image         string
		repos         []Repo
		expectedImage string
		expectError   bool
	}{
		{
			name:  "successful mapping with result",
			image: "nginx",
			repos: []Repo{
				{
					Name:        "nginx",
					CatalogTier: "APPLICATION",
					Aliases:     []string{},
				},
			},
			expectedImage: "cgr.dev/chainguard/nginx",
			expectError:   false,
		},
		{
			name:  "no results found",
			image: "nonexistent",
			repos: []Repo{
				{
					Name:        "redis",
					CatalogTier: "APPLICATION",
					Aliases:     []string{},
				},
			},
			expectError: true,
		},
		{
			name:  "multiple results returns first",
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
			expectedImage: "cgr.dev/chainguard/nginx",
			expectError:   false,
		},
		{
			name:  "image with tag",
			image: "nginx:1.25",
			repos: []Repo{
				{
					Name:        "nginx",
					CatalogTier: "APPLICATION",
					ActiveTags:  []string{"latest", "1.25", "1.26"},
					Aliases:     []string{},
				},
			},
			expectedImage: "cgr.dev/chainguard/nginx:1.25",
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &mapper{
				repos:    tc.repos,
				repoName: "cgr.dev/chainguard",
			}

			result, err := MapImage(m, tc.image)

			if tc.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.String() != tc.expectedImage {
				t.Errorf("expected %s, got %s", tc.expectedImage, result.String())
			}
		})
	}
}

func TestMapImageInvalidImage(t *testing.T) {
	m := &mapper{
		repos: []Repo{},
	}

	_, err := MapImage(m, "invalid::image")
	if err == nil {
		t.Error("expected error for invalid image reference")
	}
}

func TestMapImageMapperError(t *testing.T) {
	m := &errorMapper{err: errors.New("mapper error")}

	_, err := MapImage(m, "nginx")
	if err == nil {
		t.Error("expected error from mapper")
	}
}

type errorMapper struct {
	err error
}

func (m *errorMapper) Map(image string) (*Mapping, error) {
	return nil, m.err
}

func TestMapperIntegration(t *testing.T) {
	if v := os.Getenv("IMAGE_MAPPER_RUN_INTEGRATION_TESTS"); v == "" {
		t.Skip()
	}

	testCases := map[string][]string{
		"atmoz/sftp:alpine": {
			"cgr.dev/chainguard/atmoz-sftp",
			"cgr.dev/chainguard/atmoz-sftp-fips",
		},
		"busybox:1.35.0": {
			"cgr.dev/chainguard/busybox",
			"cgr.dev/chainguard/busybox-fips",
		},
		"coredns/coredns:1.11.3": {
			"cgr.dev/chainguard/coredns",
			"cgr.dev/chainguard/coredns-fips",
		},
		"curlimages/curl:7.85.0": {
			"cgr.dev/chainguard/curl",
			"cgr.dev/chainguard/curl-fips",
		},
		"ghcr.io/cloudnative-pg/cloudnative-pg:v1.24.4": {
			"cgr.dev/chainguard/cloudnative-pg",
			"cgr.dev/chainguard/cloudnative-pg-fips",
		},
		"ghcr.io/cloudnative-pg/pgbouncer:1.23.0": {
			"cgr.dev/chainguard/pgbouncer",
			"cgr.dev/chainguard/pgbouncer-fips",
			"cgr.dev/chainguard/pgbouncer-iamguarded",
			"cgr.dev/chainguard/pgbouncer-iamguarded-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-cloudformation:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-cloudformation",
			"cgr.dev/chainguard/crossplane-aws-cloudformation-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-cloudfront:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-cloudfront",
			"cgr.dev/chainguard/crossplane-aws-cloudfront-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-dynamodb:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-dynamodb",
			"cgr.dev/chainguard/crossplane-aws-dynamodb-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-ec2:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-ec2",
			"cgr.dev/chainguard/crossplane-aws-ec2-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-eks:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-eks",
			"cgr.dev/chainguard/crossplane-aws-eks-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-elasticache:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-elasticache",
			"cgr.dev/chainguard/crossplane-aws-elasticache-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-firehose:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-firehose",
			"cgr.dev/chainguard/crossplane-aws-firehose-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-iam:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-iam",
			"cgr.dev/chainguard/crossplane-aws-iam-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-kinesis:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-kinesis",
			"cgr.dev/chainguard/crossplane-aws-kinesis-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-kms:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-kms",
			"cgr.dev/chainguard/crossplane-aws-kms-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-lambda:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-lambda",
			"cgr.dev/chainguard/crossplane-aws-lambda-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-memorydb:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-memorydb",
			"cgr.dev/chainguard/crossplane-aws-memorydb-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-rds:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-rds",
			"cgr.dev/chainguard/crossplane-aws-rds-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-route53:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-route53",
			"cgr.dev/chainguard/crossplane-aws-route53-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-s3:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-s3",
			"cgr.dev/chainguard/crossplane-aws-s3-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-sns:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-sns",
			"cgr.dev/chainguard/crossplane-aws-sns-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-sqs:v1.20.1": {
			"cgr.dev/chainguard/crossplane-aws-sqs",
			"cgr.dev/chainguard/crossplane-aws-sqs-fips",
		},
		"ghcr.io/crossplane-contrib/provider-family-aws:v1.21.1": {
			"cgr.dev/chainguard/crossplane-aws",
			"cgr.dev/chainguard/crossplane-aws-fips",
		},
		"ghcr.io/fluxcd/flux-cli:v2.7.5": {
			"cgr.dev/chainguard/flux",
			"cgr.dev/chainguard/flux-fips",
		},
		"ghcr.io/fluxcd/helm-controller:v1.4.5": {
			"cgr.dev/chainguard/flux-helm-controller",
			"cgr.dev/chainguard/flux-helm-controller-fips",
		},
		"ghcr.io/fluxcd/image-automation-controller:v1.0.4": {
			"cgr.dev/chainguard/flux-image-automation-controller",
			"cgr.dev/chainguard/flux-image-automation-controller-fips",
		},
		"ghcr.io/fluxcd/image-reflector-controller:v1.0.4": {
			"cgr.dev/chainguard/flux-image-reflector-controller",
			"cgr.dev/chainguard/flux-image-reflector-controller-fips",
		},
		"ghcr.io/fluxcd/kustomize-controller:v1.7.3": {
			"cgr.dev/chainguard/flux-kustomize-controller",
			"cgr.dev/chainguard/flux-kustomize-controller-fips",
		},
		"ghcr.io/fluxcd/notification-controller:v1.7.5": {
			"cgr.dev/chainguard/flux-notification-controller",
			"cgr.dev/chainguard/flux-notification-controller-fips",
		},
		"ghcr.io/fluxcd/source-controller:v1.7.4": {
			"cgr.dev/chainguard/flux-source-controller",
			"cgr.dev/chainguard/flux-source-controller-fips",
		},
		"hashicorp/vault-csi-provider:1.4.0": {
			"cgr.dev/chainguard/vault-csi-provider",
			"cgr.dev/chainguard/vault-csi-provider-fips",
		},
		"hashicorp/vault:1.14.0": {
			"cgr.dev/chainguard/vault",
			"cgr.dev/chainguard/vault-fips",
		},
		"hashicorp/vault-k8s:1.14.0": {
			"cgr.dev/chainguard/vault-k8s",
			"cgr.dev/chainguard/vault-k8s-fips",
		},
		"influxdb:2.7.4-alpine": {
			"cgr.dev/chainguard/influxdb",
			"cgr.dev/chainguard/influxdb-iamguarded",
		},
		"oliver006/redis_exporter:v1.45.0-alpine": {
			"cgr.dev/chainguard/prometheus-redis-exporter",
			"cgr.dev/chainguard/prometheus-redis-exporter-fips",
		},
		"opensearchproject/opensearch-dashboards:2.19.1": {
			"cgr.dev/chainguard/opensearch-dashboards",
			"cgr.dev/chainguard/opensearch-dashboards-fips",
		},
		"opensearchproject/opensearch-operator:2.7.0": {
			"cgr.dev/chainguard/opensearch-k8s-operator",
		},
		"opensearchproject/opensearch:2.19.1": {
			"cgr.dev/chainguard/opensearch",
		},
		"percona/haproxy:2.8.5": {
			"cgr.dev/chainguard/haproxy",
			"cgr.dev/chainguard/haproxy-fips",
			"cgr.dev/chainguard/haproxy-iamguarded",
			"cgr.dev/chainguard/haproxy-iamguarded-fips",
		},
		"prom/mysqld-exporter:v0.16.0": {
			"cgr.dev/chainguard/prometheus-mysqld-exporter",
		},
		"prom/statsd-exporter:v0.26.1": {
			"cgr.dev/chainguard/prometheus-statsd-exporter",
			"cgr.dev/chainguard/prometheus-statsd-exporter-fips",
		},
		"quay.io/argoproj/argocd:v3.2.1": {
			"cgr.dev/chainguard/argocd",
			"cgr.dev/chainguard/argocd-fips",
			"cgr.dev/chainguard/argocd-iamguarded",
			"cgr.dev/chainguard/argocd-iamguarded-fips",
		},
		"quay.io/argoproj/argocli:latest": {
			"cgr.dev/chainguard/argo-cli",
			"cgr.dev/chainguard/argo-cli-fips",
		},
		"quay.io/argoproj/argoexec:latest": {
			"cgr.dev/chainguard/argo-exec",
			"cgr.dev/chainguard/argo-exec-fips",
		},
		"quay.io/argoproj/argo-events:latest": {
			"cgr.dev/chainguard/argo-events",
			"cgr.dev/chainguard/argo-events-fips",
		},
		"quay.io/argoproj/workflow-controller:latest": {
			"cgr.dev/chainguard/argo-workflowcontroller",
			"cgr.dev/chainguard/argo-workflowcontroller-fips",
		},
		"quay.io/jetstack/cert-manager-acmesolver:v1.15.2": {
			"cgr.dev/chainguard/cert-manager-acmesolver",
			"cgr.dev/chainguard/cert-manager-acmesolver-fips",
			"cgr.dev/chainguard/cert-manager-acmesolver-iamguarded",
			"cgr.dev/chainguard/cert-manager-acmesolver-iamguarded-fips",
		},
		"quay.io/jetstack/cert-manager-cainjector:v1.15.2": {
			"cgr.dev/chainguard/cert-manager-cainjector",
			"cgr.dev/chainguard/cert-manager-cainjector-fips",
			"cgr.dev/chainguard/cert-manager-cainjector-iamguarded",
			"cgr.dev/chainguard/cert-manager-cainjector-iamguarded-fips",
		},
		"quay.io/jetstack/cert-manager-controller:v1.15.2": {
			"cgr.dev/chainguard/cert-manager-controller",
			"cgr.dev/chainguard/cert-manager-controller-fips",
			"cgr.dev/chainguard/cert-manager-controller-iamguarded",
			"cgr.dev/chainguard/cert-manager-controller-iamguarded-fips",
		},
		"quay.io/jetstack/cert-manager-startupapicheck:v1.15.2": {
			"cgr.dev/chainguard/cert-manager-startupapicheck",
			"cgr.dev/chainguard/cert-manager-startupapicheck-fips",
		},
		"quay.io/jetstack/cert-manager-webhook:v1.15.2": {
			"cgr.dev/chainguard/cert-manager-webhook",
			"cgr.dev/chainguard/cert-manager-webhook-fips",
			"cgr.dev/chainguard/cert-manager-webhook-iamguarded",
			"cgr.dev/chainguard/cert-manager-webhook-iamguarded-fips",
		},
		"quay.io/jetstack/cmctl:v2.4.0": {
			"cgr.dev/chainguard/cert-manager-cmctl",
			"cgr.dev/chainguard/cert-manager-cmctl-fips",
		},
		"quay.io/jetstack/trust-manager:v0.12.0": {
			"cgr.dev/chainguard/trust-manager",
			"cgr.dev/chainguard/trust-manager-fips",
		},
		"quay.io/minio/mc:RELEASE.2025-08-13T08-35-41Z-cpuv1": {
			"cgr.dev/chainguard/minio-client",
			"cgr.dev/chainguard/minio-client-fips",
		},
		"quay.io/minio/minio:RELEASE.2024-10-02T17-50-41Z": {
			"cgr.dev/chainguard/minio",
			"cgr.dev/chainguard/minio-fips",
			"cgr.dev/chainguard/minio-iamguarded",
			"cgr.dev/chainguard/minio-iamguarded-fips",
		},
		"quay.io/minio/operator:v6.0.4": {
			"cgr.dev/chainguard/minio-operator",
			"cgr.dev/chainguard/minio-operator-fips",
		},
		"quay.io/minio/operator-sidecar:v6.0.4": {
			"cgr.dev/chainguard/minio-operator-sidecar",
			"cgr.dev/chainguard/minio-operator-sidecar-fips",
		},
		"quay.io/mongodb/mongodb-kubernetes-operator-version-upgrade-post-start-hook:1.0.9": {
			"cgr.dev/chainguard/mongodb-kubernetes-operator-version-upgrade-post-start-hook",
			"cgr.dev/chainguard/mongodb-kubernetes-operator-version-upgrade-post-start-hook-fips",
		},
		"quay.io/mongodb/mongodb-kubernetes-operator:0.12.0": {
			"cgr.dev/chainguard/mongodb-kubernetes-operator",
			"cgr.dev/chainguard/mongodb-kubernetes-operator-fips",
		},
		"quay.io/prometheus/pushgateway:v1.9.0": {
			"cgr.dev/chainguard/prometheus-pushgateway",
			"cgr.dev/chainguard/prometheus-pushgateway-fips",
			"cgr.dev/chainguard/prometheus-pushgateway-iamguarded",
			"cgr.dev/chainguard/prometheus-pushgateway-iamguarded-fips",
		},
		"registry.k8s.io/sig-storage/csi-attacher:v4.6.1": {
			"cgr.dev/chainguard/kubernetes-csi-external-attacher",
		},
		"registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.11.1": {
			"cgr.dev/chainguard/kubernetes-csi-node-driver-registrar",
		},
		"registry.k8s.io/sig-storage/livenessprobe:v2.13.1": {
			"cgr.dev/chainguard/kubernetes-csi-livenessprobe",
		},
		"registry.k8s.io/sig-storage/nfsplugin:v4.8.0": {
			"cgr.dev/chainguard/kubernetes-csi-driver-nfs",
		},
		"registry.k8s.io/sig-storage/snapshot-controller:v8.0.1": {
			"cgr.dev/chainguard/kubernetes-csi-external-snapshot-controller",
			"cgr.dev/chainguard/kubernetes-csi-external-snapshotter",
		},
		"thingsboard/tb-js-executor:3.9.1": {
			"cgr.dev/chainguard/thingsboard-tb-js-executor",
		},
		"thingsboard/tb-mqtt-transport:3.5.1": {
			"cgr.dev/chainguard/thingsboard-tb-mqtt-transport",
		},
		"thingsboard/tb-node:3.5.1": {
			"cgr.dev/chainguard/thingsboard-tb-node",
		},
		"thingsboard/tb-web-ui:3.5.1": {
			"cgr.dev/chainguard/thingsboard-tb-web-ui",
		},
		"valkey/valkey:7.2.5-alpine": {
			"cgr.dev/chainguard/valkey",
			"cgr.dev/chainguard/valkey-fips",
			"cgr.dev/chainguard/valkey-iamguarded",
			"cgr.dev/chainguard/valkey-iamguarded-fips",
		},
	}

	ctx := t.Context()
	m, err := NewMapper(ctx)
	if err != nil {
		t.Fatalf("unexpected error creating mapper: %s", err)
	}

	// stripTag removes the tag portion from an image reference
	stripTag := func(img string) string {
		if idx := strings.Index(img, ":"); idx != -1 {
			return img[:idx]
		}
		return img
	}

	for img, wantResults := range testCases {
		t.Run(img, func(t *testing.T) {
			got, err := m.Map(img)
			if err != nil {
				t.Errorf("unexpected error mapping %s: %s", img, err)
			}

			want := &Mapping{
				Image:   img,
				Results: wantResults,
			}

			// Compare image references without tags since tags are dynamic
			// Sort based on the stripped image names (without tags)
			opts := cmp.Options{
				cmpopts.AcyclicTransformer("StripTags", func(s string) string {
					return stripTag(s)
				}),
				cmpopts.SortSlices(func(a, b string) bool {
					return strings.Compare(stripTag(a), stripTag(b)) < 0
				}),
			}

			if diff := cmp.Diff(want, got, opts); diff != "" {
				t.Errorf("unexpected mapping for %s:\n%s", img, diff)
			}
		})
	}

}
