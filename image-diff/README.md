# `image-diff`

This demonstrates how to query diffs of Chainguard Images.

Along with Tag History, you can use this to show the evolution of an image over time, or the difference between packages in a `:latest` and a `:latest-dev` image tag.

### Usage

```sh
reponame=go
left=sha256:a62aded9da72d0f4ad2f6eb751f3ce3fff2f5d0d30d93dcd245c0cd650d5028a  # :latest
right=sha256:9d49f4b2d67988d5345419a35533762e72eaaa8162d4b43a1e3d41869d1f845e # :latest-dev
go run ./cmd/app $reponame $previous $current
```
