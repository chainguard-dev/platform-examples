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
				repos:     tc.repos,
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
		repos: repos,
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
		repos: repos,
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
	m := &Mapper{
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
				Results: []string{"nginx", "prod-nginx"},
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
				Results: []string{"redis", "redis-prod"},
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
				Results: []string{"postgres", "postgres-prod"},
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
				Results: []string{"mysql", "mysql-new"},
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
				Results: []string{"node", "node-staging"},
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
				Results: []string{"python", "python-slim"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &Mapper{
				repos:     tc.repos,
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

	m := &Mapper{
		repos:     repos,
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
		Results: []string{"web-server"},
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

	m := &Mapper{
		repos:     repos,
		ignoreFns: []IgnoreFn{}, // No ignore functions
	}

	result, err := m.Map("nginx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should get all matching repos when no ignore functions are set
	expected := &Mapping{
		Image:   "nginx",
		Results: []string{"nginx", "nginx-dev", "nginx-test"},
	}

	opts := cmpopts.SortSlices(func(a, b string) bool {
		return strings.Compare(a, b) < 0
	})

	if diff := cmp.Diff(expected, result, opts); diff != "" {
		t.Errorf("mapping mismatch (-want +got):\n%s", diff)
	}
}

func TestMapperIntegration(t *testing.T) {
	if v := os.Getenv("IMAGE_MAPPER_RUN_INTEGRATION_TESTS"); v == "" {
		t.Skip()
	}

	testCases := map[string][]string{
		"atmoz/sftp:alpine": {
			"atmoz-sftp",
			"atmoz-sftp-fips",
		},
		"busybox:1.35.0": {
			"busybox",
			"busybox-fips",
		},
		"coredns/coredns:1.11.3": {
			"coredns",
			"coredns-fips",
		},
		"curlimages/curl:7.85.0": {
			"curl",
			"curl-fips",
		},
		"ghcr.io/cloudnative-pg/cloudnative-pg:v1.24.4": {
			"cloudnative-pg",
			"cloudnative-pg-fips",
		},
		"ghcr.io/cloudnative-pg/pgbouncer:1.23.0": {
			"pgbouncer",
			"pgbouncer-fips",
			"pgbouncer-iamguarded",
			"pgbouncer-iamguarded-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-cloudformation:v1.20.1": {
			"crossplane-aws-cloudformation",
			"crossplane-aws-cloudformation-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-cloudfront:v1.20.1": {
			"crossplane-aws-cloudfront",
			"crossplane-aws-cloudfront-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-dynamodb:v1.20.1": {
			"crossplane-aws-dynamodb",
			"crossplane-aws-dynamodb-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-ec2:v1.20.1": {
			"crossplane-aws-ec2",
			"crossplane-aws-ec2-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-eks:v1.20.1": {
			"crossplane-aws-eks",
			"crossplane-aws-eks-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-firehose:v1.20.1": {
			"crossplane-aws-firehose",
			"crossplane-aws-firehose-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-iam:v1.20.1": {
			"crossplane-aws-iam",
			"crossplane-aws-iam-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-kinesis:v1.20.1": {
			"crossplane-aws-kinesis",
			"crossplane-aws-kinesis-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-kms:v1.20.1": {
			"crossplane-aws-kms",
			"crossplane-aws-kms-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-lambda:v1.20.1": {
			"crossplane-aws-lambda",
			"crossplane-aws-lambda-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-rds:v1.20.1": {
			"crossplane-aws-rds",
			"crossplane-aws-rds-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-route53:v1.20.1": {
			"crossplane-aws-route53",
			"crossplane-aws-route53-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-s3:v1.20.1": {
			"crossplane-aws-s3",
			"crossplane-aws-s3-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-sns:v1.20.1": {
			"crossplane-aws-sns",
			"crossplane-aws-sns-fips",
		},
		"ghcr.io/crossplane-contrib/provider-aws-sqs:v1.20.1": {
			"crossplane-aws-sqs",
			"crossplane-aws-sqs-fips",
		},
		"ghcr.io/crossplane-contrib/provider-family-aws:v1.21.1": {
			"crossplane-aws",
			"crossplane-aws-fips",
		},
		"ghcr.io/fluxcd/flux-cli:v2.7.5": {
			"flux",
			"flux-fips",
		},
		"ghcr.io/fluxcd/helm-controller:v1.4.5": {
			"flux-helm-controller",
			"flux-helm-controller-fips",
		},
		"ghcr.io/fluxcd/image-automation-controller:v1.0.4": {
			"flux-image-automation-controller",
			"flux-image-automation-controller-fips",
		},
		"ghcr.io/fluxcd/image-reflector-controller:v1.0.4": {
			"flux-image-reflector-controller",
			"flux-image-reflector-controller-fips",
		},
		"ghcr.io/fluxcd/kustomize-controller:v1.7.3": {
			"flux-kustomize-controller",
			"flux-kustomize-controller-fips",
		},
		"ghcr.io/fluxcd/notification-controller:v1.7.5": {
			"flux-notification-controller",
			"flux-notification-controller-fips",
		},
		"ghcr.io/fluxcd/source-controller:v1.7.4": {
			"flux-source-controller",
			"flux-source-controller-fips",
		},
		"hashicorp/vault-csi-provider:1.4.0": {
			"vault-csi-provider",
			"vault-csi-provider-fips",
		},
		"hashicorp/vault:1.14.0": {
			"vault",
			"vault-fips",
		},
		"hashicorp/vault-k8s:1.14.0": {
			"vault-k8s",
			"vault-k8s-fips",
		},
		"influxdb:2.7.4-alpine": {
			"influxdb",
			"influxdb-iamguarded",
		},
		"oliver006/redis_exporter:v1.45.0-alpine": {
			"prometheus-redis-exporter",
			"prometheus-redis-exporter-fips",
		},
		"opensearchproject/opensearch-dashboards:2.19.1": {
			"opensearch-dashboards",
			"opensearch-dashboards-fips",
		},
		"opensearchproject/opensearch-operator:2.7.0": {
			"opensearch-k8s-operator",
		},
		"opensearchproject/opensearch:2.19.1": {
			"opensearch",
		},
		"percona/haproxy:2.8.5": {
			"haproxy",
			"haproxy-fips",
			"haproxy-iamguarded",
			"haproxy-iamguarded-fips",
		},
		"prom/mysqld-exporter:v0.16.0": {
			"prometheus-mysqld-exporter",
		},
		"prom/statsd-exporter:v0.26.1": {
			"prometheus-statsd-exporter",
			"prometheus-statsd-exporter-fips",
		},
		"quay.io/argoproj/argocd:v3.2.1": {
			"argocd",
			"argocd-fips",
			"argocd-iamguarded",
			"argocd-iamguarded-fips",
			"argocd-repo-server",
			"argocd-repo-server-fips",
		},
		"quay.io/argoproj/argocli:latest": {
			"argo-cli",
			"argo-cli-fips",
		},
		"quay.io/argoproj/argoexec:latest": {
			"argo-exec",
			"argo-exec-fips",
		},
		"quay.io/argoproj/argo-events:latest": {
			"argo-events",
			"argo-events-fips",
		},
		"quay.io/argoproj/workflow-controller:latest": {
			"argo-workflowcontroller",
			"argo-workflowcontroller-fips",
		},
		"quay.io/jetstack/cert-manager-acmesolver:v1.15.2": {
			"cert-manager-acmesolver",
			"cert-manager-acmesolver-fips",
			"cert-manager-acmesolver-iamguarded",
			"cert-manager-acmesolver-iamguarded-fips",
		},
		"quay.io/jetstack/cert-manager-cainjector:v1.15.2": {
			"cert-manager-cainjector",
			"cert-manager-cainjector-fips",
			"cert-manager-cainjector-iamguarded",
			"cert-manager-cainjector-iamguarded-fips",
		},
		"quay.io/jetstack/cert-manager-controller:v1.15.2": {
			"cert-manager-controller",
			"cert-manager-controller-fips",
			"cert-manager-controller-iamguarded",
			"cert-manager-controller-iamguarded-fips",
		},
		"quay.io/jetstack/cert-manager-startupapicheck:v1.15.2": {
			"cert-manager-startupapicheck",
			"cert-manager-startupapicheck-fips",
		},
		"quay.io/jetstack/cert-manager-webhook:v1.15.2": {
			"cert-manager-webhook",
			"cert-manager-webhook-fips",
			"cert-manager-webhook-iamguarded",
			"cert-manager-webhook-iamguarded-fips",
		},
		"quay.io/jetstack/cmctl:v2.4.0": {
			"cert-manager-cmctl",
			"cert-manager-cmctl-fips",
		},
		"quay.io/jetstack/trust-manager:v0.12.0": {
			"trust-manager",
			"trust-manager-fips",
		},
		"quay.io/minio/mc:RELEASE.2025-08-13T08-35-41Z-cpuv1": {
			"minio-client",
			"minio-client-fips",
		},
		"quay.io/minio/minio:RELEASE.2024-10-02T17-50-41Z": {
			"minio",
			"minio-fips",
			"minio-iamguarded",
			"minio-iamguarded-fips",
		},
		"quay.io/minio/operator:v6.0.4": {
			"minio-operator",
			"minio-operator-fips",
		},
		"quay.io/minio/operator-sidecar:v6.0.4": {
			"minio-operator-sidecar",
			"minio-operator-sidecar-fips",
		},
		"quay.io/mongodb/mongodb-kubernetes-operator-version-upgrade-post-start-hook:1.0.9": {
			"mongodb-kubernetes-operator-version-upgrade-post-start-hook",
			"mongodb-kubernetes-operator-version-upgrade-post-start-hook-fips",
		},
		"quay.io/mongodb/mongodb-kubernetes-operator:0.12.0": {
			"mongodb-kubernetes-operator",
			"mongodb-kubernetes-operator-fips",
		},
		"quay.io/prometheus/pushgateway:v1.9.0": {
			"prometheus-pushgateway",
			"prometheus-pushgateway-fips",
			"prometheus-pushgateway-iamguarded",
			"prometheus-pushgateway-iamguarded-fips",
		},
		"registry.k8s.io/sig-storage/csi-attacher:v4.6.1": {
			"kubernetes-csi-external-attacher",
		},
		"registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.11.1": {
			"kubernetes-csi-node-driver-registrar",
		},
		"registry.k8s.io/sig-storage/livenessprobe:v2.13.1": {
			"kubernetes-csi-livenessprobe",
		},
		"registry.k8s.io/sig-storage/nfsplugin:v4.8.0": {
			"kubernetes-csi-driver-nfs",
		},
		"registry.k8s.io/sig-storage/snapshot-controller:v8.0.1": {
			"kubernetes-csi-external-snapshot-controller",
			"kubernetes-csi-external-snapshotter",
		},
		"thingsboard/tb-js-executor:3.9.1": {
			"thingsboard-tb-js-executor",
		},
		"thingsboard/tb-mqtt-transport:3.5.1": {
			"thingsboard-tb-mqtt-transport",
		},
		"thingsboard/tb-node:3.5.1": {
			"thingsboard-tb-node",
		},
		"thingsboard/tb-web-ui:3.5.1": {
			"thingsboard-tb-web-ui",
		},
		"valkey/valkey:7.2.5-alpine": {
			"valkey",
			"valkey-fips",
			"valkey-iamguarded",
			"valkey-iamguarded-fips",
		},
	}

	ctx := t.Context()
	m, err := NewMapper(ctx)
	if err != nil {
		t.Fatalf("unexpected error creating mapper: %s", err)
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
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("unexpected mapping for %s:\n%s", img, diff)
			}
		})
	}

}
