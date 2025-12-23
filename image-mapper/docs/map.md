# Map

The `map` command maps image references to their Chainguard equivalents.

## Usage

You can pass images on the command line. 

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
```

## Options

### Output

Configure the output format with the `-o` flag. Supported formats are: `csv`,
`json` and `text`.

```
$ ./image-mapper map ghcr.io/stakater/reloader:v1.4.1 registry.k8s.io/sig-storage/livenessprobe:v2.13.1 -o json | jq -r .
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
$ ./image-mapper map ghcr.io/stakater/reloader:v1.4.1 registry.k8s.io/sig-storage/livenessprobe:v2.13.1 -o csv
ghcr.io/stakater/reloader:v1.4.1,[cgr.dev/chainguard/stakater-reloader-fips:v1.4.12 cgr.dev/chainguard/stakater-reloader:v1.4.12]
registry.k8s.io/sig-storage/livenessprobe:v2.13.1,[cgr.dev/chainguard/kubernetes-csi-livenessprobe:v2.17.0]
```

### Ignore Tiers (i.e FIPS)

The output will map both FIPS and non-FIPS variants. You can exclude FIPS with
the `--ignore-tiers` flag.

```
$ ./image-mapper map prom/prometheus
prom/prometheus -> cgr.dev/chainguard/prometheus-fips:latest
prom/prometheus -> cgr.dev/chainguard/prometheus-iamguarded-fips:latest
prom/prometheus -> cgr.dev/chainguard/prometheus-iamguarded:latest
prom/prometheus -> cgr.dev/chainguard/prometheus:latest

$ ./image-mapper map prom/prometheus --ignore-tiers=FIPS
prom/prometheus -> cgr.dev/chainguard/prometheus-iamguarded:latest
prom/prometheus -> cgr.dev/chainguard/prometheus:latest
```

### Ignore Iamguarded

The mapper will also return matches for our `-iamguarded` images. These images
are designed specifically to work with Chainguard's Helm charts. If you aren't
interested in using our charts, you can exclude those matches with
`--ignore-iamguarded`.

```
$ ./image-mapper map prom/prometheus --ignore-iamguarded
prom/prometheus -> cgr.dev/chainguard/prometheus-fips:latest
prom/prometheus -> cgr.dev/chainguard/prometheus:latest
```
