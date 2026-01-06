# Map Dockerfile

Map images references in a Dockerfile to Chainguard images.

## How It Works

The `dockerfile` subcommand maps any image references it finds in  `FROM <image>`,
`COPY --from=<image>` or `RUN --mount-type=bind,from=<image>` directives to
Chainguard.

It will map images to `-dev` tags because they are more likely to work out of
the box as drop in replacements.

## Basic Usage

Given a `Dockerfile` like this:

```
FROM python:3.13 AS python

WORKDIR /app

COPY run.py run.py

ENTRYPOINT ["python", "/app/run.py"]
```

Use the `dockerfile` subcommand to map it to Chainguard images. It returns the
result to stdout.

```
$ ./image-mapper map dockerfile Dockerfile
FROM cgr.dev/chainguard/python:3.13-dev

WORKDIR /app

COPY run.py run.py

ENTRYPOINT ["python", "/app/run.py"]
```

You can also provide the Dockerfile via stdin:

```
$ cat Dockerfile | ./image-mapper map dockerfile -
```

## Repository Prefix

Use the `--repository` flag to replace `cgr.dev/chainguard` with a custom
repository.

```
$ ./image-mapper map dockerfile Dockerfile --repository=registry.internal/cgr
FROM registry.internal/cgr/python:3.13-dev

WORKDIR /app

COPY run.py run.py

ENTRYPOINT ["python", "/app/run.py"]
```

## Known Limitations

There are a few rough edges that haven't been smoothed out yet.

### Args

The mapper supports resolving arguments to figure out which images they refer
to but it isn't clever enough to go back and update those arguments.

For instance, a file like this:

```
ARG REGISTRY=docker.io
ARG IMAGE=library/python
ARG TAG=3.13
FROM ${REGISTRY}/${IMAGE}:${TAG}
```

Would become:

```
ARG REGISTRY=docker.io
ARG IMAGE=library/python
ARG TAG=3.13
FROM cgr.dev/chainguard/python:3.13-dev
```

### Multi Line Directives

If it updates an image reference in a multi line directive then it will squash
the directive into one line.

For instance, lines like this:

```
RUN --mount=type=bind,from=ubuntu,target=/bin/cat \
     cat run.py
```

Would become:

```
RUN --mount=type=bind,from=cgr.dev/chainguard/chainguard-base:latest,target=/bin/cat cat run.py
```
