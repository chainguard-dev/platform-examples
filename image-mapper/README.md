# image-mapper

A tool for matching non-Chainguard images to their Chainguard equivalents.

## Build

Build the tool.

```
$ go build -o image-mapper .
```

You can also build and run the tool with Docker.

```
# Build the image
docker build -t image-mapper .

# Run for an individual image
docker run -it --rm image-mapper map ghcr.io/stakater/reloader:v1.4.1

# Or, pass a list of images from a text file
docker run -i --rm image-mapper -- map - < images.txt
```

## Basic Usage

### Map

The `map` command maps images provided on the command line.

```
$ ./image-mapper map ghcr.io/stakater/reloader:v1.4.1 registry.k8s.io/sig-storage/livenessprobe:v2.13.1
ghcr.io/stakater/reloader:v1.4.1 -> cgr.dev/chainguard/stakater-reloader-fips:v1.4.12
ghcr.io/stakater/reloader:v1.4.1 -> cgr.dev/chainguard/stakater-reloader:v1.4.12
registry.k8s.io/sig-storage/livenessprobe:v2.13.1 -> cgr.dev/chainguard/kubernetes-csi-livenessprobe:v2.17.0
```

You can also provide a list of images (one image per line) via stdin when the first
argument is `-`.

```
$ cat ./images.txt | ./image-mapper map -
ghcr.io/stakater/reloader:v1.4.1 -> cgr.dev/chainguard/stakater-reloader-fips:v1.4.12
ghcr.io/stakater/reloader:v1.4.1 -> cgr.dev/chainguard/stakater-reloader:v1.4.12
registry.k8s.io/sig-storage/livenessprobe:v2.13.1 -> cgr.dev/chainguard/kubernetes-csi-livenessprobe:v2.17.0
```

You'll notice that the mapper increments the tag to the closest version
supported by Chainguard. To benefit from continued CVE remediation, it's
important, where possible, to use tags that are being actively maintained.

Refer to [this page](./docs/map.md) for more details.

### Helm

The `helm-chart` and `helm-values` subcommands extract image related values and
map them to Chainguard.

```
$ ./image-mapper map helm-chart argocd/argo-cd
redis-ha:
    image:
        repository: cgr.dev/chainguard/redis # Original: ecr-public.aws.com/docker/library/redis
        tag: 8.2.2 # Original: 8.2.2-alpine
    configmapTest:
        image:
            repository: cgr.dev/chainguard/shellcheck # Original: koalaman/shellcheck
            tag: v0.11.0-dev # Original: v0.10.0
    haproxy:
        image:
            repository: cgr.dev/chainguard/haproxy # Original: ecr-public.aws.com/docker/library/haproxy
    exporter:
        image: cgr.dev/chainguard/prometheus-redis-exporter # Original: ghcr.io/oliver006/redis_exporter
global:
    image:
        repository: cgr.dev/chainguard/argocd # Original: quay.io/argoproj/argocd
...
```

```
$ helm show values argocd/argo-cd | ./image-mapper map helm-values -
global:
    image:
        repository: cgr.dev/chainguard/argocd # Original: quay.io/argoproj/argocd
dex:
    image:
        repository: cgr.dev/chainguard/dex # Original: ghcr.io/dexidp/dex
...
```

These commands provide values overrides that you can pass to `helm install`.

Refer to [this page](./docs/map_helm.md) for more details.

## Development

You can run integration tests against the actual catalog endpoint by setting
`IMAGE_MAPPER_RUN_INTEGRATION_TESTS=1`:

```
IMAGE_MAPPER_RUN_INTEGRATION_TESTS=1 go test ./...
```

This identifies regressions in the mapping logic or the catalog data by
recording known matches.
