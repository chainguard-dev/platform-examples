# Map Helm

Streamline the process of migrating Helm charts to Chainguard by
extracting the image related values from a chart and mapping them to
Chainguard images.

## Charts

The `helm-chart` subcommand extracts all the image related values from a Helm
chart, as well as its dependencies and subcharts.

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
dex:
    image:
        repository: cgr.dev/chainguard/dex # Original: ghcr.io/dexidp/dex
redis:
    image:
        repository: cgr.dev/chainguard/redis # Original: ecr-public.aws.com/docker/library/redis
        tag: 8.2.2 # Original: 8.2.2-alpine
    exporter:
        image:
            repository: cgr.dev/chainguard/prometheus-redis-exporter # Original: ghcr.io/oliver006/redis_exporter
server:
    extensions:
        image:
            repository: cgr.dev/chainguard/argocd-extension-installer # Original: quay.io/argoprojlabs/argocd-extension-installer
```

This command doesn't require that `helm` is available locally but if you are 
using a chart reference of the form `<repo>/<chart>` (i.e `argocd/argo-cd`)
then you must have added the repository with `helm repo add`.

Alternatively, you can specify the chart repository and/or the chart version
directly. This will pull the chart directly, with no dependency on local
configuration at all.

```
$ ./image-mapper map helm-chart argo-cd \
    --chart-repo=https://argoproj.github.io/argo-helm \
    --chart-version=9.1.0
```

## Values

The `helm-values` subcommand extracts all the image related values from a values
file.

This will only identify values that are explicitly listed in the file and may
not include all the values that are provided by subcharts or dependencies.

```
$ helm show values argocd/argo-cd | ./image-mapper map helm-values -
global:
    image:
        repository: cgr.dev/chainguard/argocd # Original: quay.io/argoproj/argocd
dex:
    image:
        repository: cgr.dev/chainguard/dex # Original: ghcr.io/dexidp/dex
redis:
    image:
        repository: cgr.dev/chainguard/redis # Original: ecr-public.aws.com/docker/library/redis
        tag: 8.2.2 # Original: 8.2.2-alpine
    exporter:
        image:
            repository: cgr.dev/chainguard/prometheus-redis-exporter # Original: ghcr.io/oliver006/redis_exporter
redis-ha:
    image:
        repository: cgr.dev/chainguard/redis # Original: ecr-public.aws.com/docker/library/redis
        tag: 8.2.2 # Original: 8.2.2-alpine
    exporter:
        image: cgr.dev/chainguard/prometheus-redis-exporter # Original: ghcr.io/oliver006/redis_exporter
    haproxy:
        image:
            repository: cgr.dev/chainguard/haproxy # Original: ecr-public.aws.com/docker/library/haproxy
server:
    extensions:
        image:
            repository: cgr.dev/chainguard/argocd-extension-installer # Original: quay.io/argoprojlabs/argocd-extension-installer
```

## Options

Both commands support a `--repository` flag which configures the repository
images are mapped to. This allows you to include your mirror or proxy URL in the
mappings.

```
$ ./image-mapper map helm-chart prometheus-community/kube-state-metrics --repository=registry.internal.mirror/cgr
image:
    registry: registry.internal.mirror # Original: registry.k8s.io
    repository: cgr/kube-state-metrics # Original: kube-state-metrics/kube-state-metrics
kubeRBACProxy:
    image:
        registry: registry.internal.mirror # Original: quay.io
        repository: cgr/kube-rbac-proxy # Original: brancz/kube-rbac-proxy
```

## Testing

You can validate whether the returned values have overridden all the images by
passing them to `helm template` and grepping for images.

```
$ ./image-mapper map helm-chart argocd/argo-cd \
    | helm template argocd/argo-cd -f - \
    | grep 'image:'
          image: cgr.dev/chainguard/argocd:v3.2.1
          image: cgr.dev/chainguard/argocd:v3.2.1
        image: cgr.dev/chainguard/argocd:v3.2.1
        image: cgr.dev/chainguard/argocd:v3.2.1
        image: cgr.dev/chainguard/argocd:v3.2.1
        image: cgr.dev/chainguard/dex:v2.44.0
        image: cgr.dev/chainguard/argocd:v3.2.1
        image: cgr.dev/chainguard/redis:8.2.2
        image: cgr.dev/chainguard/argocd:v3.2.1
        image: cgr.dev/chainguard/argocd:v3.2.1
```
