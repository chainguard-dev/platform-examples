# image-mapper

A tool for matching non-Chainguard images to their Chainguard equivalents.

## Usage

Build the tool.

```
$ go build -o image-mapper .
```

Then, provide the images to map on the command line.

```
$ ./image-mapper ghcr.io/stakater/reloader:v1.4.1 registry.k8s.io/sig-storage/livenessprobe:v2.13.1
ghcr.io/stakater/reloader:v1.4.1 -> cgr.dev/chainguard/stakater-reloader-fips:v1.4.12
ghcr.io/stakater/reloader:v1.4.1 -> cgr.dev/chainguard/stakater-reloader:v1.4.12
registry.k8s.io/sig-storage/livenessprobe:v2.13.1 -> cgr.dev/chainguard/kubernetes-csi-livenessprobe:v2.17.0
```

You'll notice that the mapper increments the tag to the closest version
supported by Chainguard. To benefit from continued CVE remediation, it's
important, where possible, to use tags that are being actively maintained.

You can also provide a list of images (one image per line) via stdin when the first
argument is `-`.

```
$ cat ./images.txt | ./image-mapper -
```

## Options

### Output

Configure the output format with the `-o` flag. Supported formats are: `csv`,
`json` and `text`.

```
$ ./image-mapper ghcr.io/stakater/reloader:v1.4.1 registry.k8s.io/sig-storage/livenessprobe:v2.13.1 -o json | jq -r .
[
  {
    "image": "ghcr.io/stakater/reloader:v1.4.1",
    "results": [
      "cgr.dev/chainguard/stakater-reloader-fips:v1.4.12",
      "cgr.dev/chainguard/stakater-reloader:v1.4.12"
    ]
  },
  {
    "image": "registry.k8s.io/sig-storage/livenessprobe:v2.13.1",
    "results": [
      "cgr.dev/chainguard/kubernetes-csi-livenessprobe:v2.17.0"
    ]
  }
]
```

```
$ ./image-mapper ghcr.io/stakater/reloader:v1.4.1 registry.k8s.io/sig-storage/livenessprobe:v2.13.1 -o csv
ghcr.io/stakater/reloader:v1.4.1,[cgr.dev/chainguard/stakater-reloader-fips:v1.4.12 cgr.dev/chainguard/stakater-reloader:v1.4.12]
registry.k8s.io/sig-storage/livenessprobe:v2.13.1,[cgr.dev/chainguard/kubernetes-csi-livenessprobe:v2.17.0]
```

### Ignore Tiers (i.e FIPS)

The output will map both FIPS and non-FIPS variants. You can exclude FIPS with
the `--ignore-tiers` flag.

```
$ ./image-mapper prom/prometheus
prom/prometheus -> cgr.dev/chainguard/prometheus-fips:latest
prom/prometheus -> cgr.dev/chainguard/prometheus-iamguarded-fips:latest
prom/prometheus -> cgr.dev/chainguard/prometheus-iamguarded:latest
prom/prometheus -> cgr.dev/chainguard/prometheus:latest

$ ./image-mapper prom/prometheus --ignore-tiers=FIPS
prom/prometheus -> cgr.dev/chainguard/prometheus-iamguarded:latest
prom/prometheus -> cgr.dev/chainguard/prometheus:latest
```

### Ignore Iamguarded

The mapper will also return matches for our `-iamguarded` images. These images
are designed specifically to work with Chainguard's Helm charts. If you aren't
interested in using our charts, you can exclude those matches with
`--ignore-iamguarded`.

```
$ ./image-mapper prom/prometheus --ignore-iamguarded
prom/prometheus -> cgr.dev/chainguard/prometheus-fips:latest
prom/prometheus -> cgr.dev/chainguard/prometheus:latest
```

## Docker

```
# Build the image
docker build -t image-mapper .

# Run for an individual image
docker run -it --rm image-mapper ghcr.io/stakater/reloader:v1.4.1

# Or, pass a list of images from a text file
docker run -i --rm image-mapper -- - < images.txt
```

## Development

You can run integration tests against the actual catalog endpoint by setting
`IMAGE_MAPPER_RUN_INTEGRATION_TESTS=1`:

```
IMAGE_MAPPER_RUN_INTEGRATION_TESTS=1 go test ./...
```

This identifies regressions in the mapping logic or the catalog data by
recording known matches.
