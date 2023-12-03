# `image-diff`

This demonstrates how to query diffs of Chainguard Images.

Along with Tag History, you can use this to show the evolution of an image over time.

### Usage


```sh
previous=sha256:0b7ee54f5b593bd9463cc56a37d16b447de4ea95492dba9cfa5ce2d732e6352c
current=sha256:c894bc454800817b1747c8a1a640ae6d86004b06190f94e791098e7e78dbbc00
go run ./cmd/app go $previous $current
```
