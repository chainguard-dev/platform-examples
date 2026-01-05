package helm

import (
	"testing"

	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/mapper"
	"github.com/google/go-cmp/cmp"
)

type mockMapper struct {
	mappings map[string][]string
}

func (m *mockMapper) Map(img string) (*mapper.Mapping, error) {
	return &mapper.Mapping{
		Image:   img,
		Results: m.mappings[img],
	}, nil
}

func TestMapValues(t *testing.T) {
	input := []byte(`
prometheus:
    image: prom/prometheus:v2.18.1
redis-example:
    exporter:
        enabled: true
        image: ghcr.io/oliver006/redis_exporter
        tag: v1.75.0
    haproxy:
        enabled: false
        image:
            repository: ecr-public.aws.com/docker/library/haproxy
    image:
        registry: ecr-public.aws.com
        repository: docker/library/redis
proxy:
  traefik:
    image:
      name: traefik
      tag: v3.6.4

global:
  revisionHistoryLimit: 3
  image:
    repository: quay.io/argoproj/argocd
    tag: ""

prometheus-example:
  admissionWebhooks:
      deployment:
          image:
              registry: "quay.io"
              repository: prometheus-operator/admission-webhook
      patch:
          image:
              registry: "ghcr.io"
              repository: jkroepke/kube-webhook-certgen
  image:
      registry: ""
      repository: prometheus-operator/prometheus-operator
      tag: ""
      sha: ""
`)

	want := []byte(`prometheus:
    image: cgr.dev/chainguard/prometheus:v2.56.0 # Original: prom/prometheus:v2.18.1
redis-example:
    exporter:
        image: cgr.dev/chainguard/prometheus-redis-exporter # Original: ghcr.io/oliver006/redis_exporter
        tag: v1.76.0 # Original: v1.75.0
    haproxy:
        image:
            repository: cgr.dev/chainguard/haproxy # Original: ecr-public.aws.com/docker/library/haproxy
    image:
        registry: cgr.dev # Original: ecr-public.aws.com
        repository: chainguard/redis # Original: docker/library/redis
proxy:
    traefik:
        image:
            name: cgr.dev/chainguard/traefik # Original: traefik
global:
    image:
        repository: cgr.dev/chainguard/argocd # Original: quay.io/argoproj/argocd
prometheus-example:
    admissionWebhooks:
        deployment:
            image:
                registry: cgr.dev # Original: quay.io
                repository: chainguard/prometheus-admission-webhook # Original: prometheus-operator/admission-webhook
        patch:
            image:
                registry: cgr.dev # Original: ghcr.io
                repository: chainguard/kube-webhook-certgen # Original: jkroepke/kube-webhook-certgen
    image:
        registry: cgr.dev # Original: 
        repository: chainguard/prometheus-operator # Original: prometheus-operator/prometheus-operator
`)

	m := &mockMapper{
		mappings: map[string][]string{
			"ecr-public.aws.com/docker/library/haproxy": {
				"cgr.dev/chainguard/haproxy:latest",
			},
			"ecr-public.aws.com/docker/library/redis": {
				"cgr.dev/chainguard/redis:latest",
			},
			"ghcr.io/jkroepke/kube-webhook-certgen": {
				"cgr.dev/chainguard/kube-webhook-certgen:latest",
			},
			"ghcr.io/oliver006/redis_exporter:v1.75.0": {
				"cgr.dev/chainguard/prometheus-redis-exporter:v1.76.0",
			},
			"quay.io/argoproj/argocd": {
				"cgr.dev/chainguard/argocd:latest",
			},
			"quay.io/prometheus-operator/admission-webhook": {
				"cgr.dev/chainguard/prometheus-admission-webhook:latest",
			},
			"prom/prometheus:v2.18.1": {
				"cgr.dev/chainguard/prometheus:v2.56.0",
			},
			"prometheus-operator/prometheus-operator": {
				"cgr.dev/chainguard/prometheus-operator:latest",
			},
			"traefik:v3.6.4": {
				"cgr.dev/chainguard/traefik:v3.6.4",
			},
		},
	}

	got, err := mapValues(m, input)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected output:\n%s", diff)
	}
}
