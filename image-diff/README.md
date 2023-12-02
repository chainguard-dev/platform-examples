# `image-diff`

This demonstrates how to query diffs of Chainguard Images.

Along with Tag History, you can use this to show the evolution of an image over time.

### Usage


```sh
previous=sha256:10fe8e11120a983bce706e054a83f1ec96505bcb26fd904ed115767f4070f3f2
current=sha256:ec687431d948ca883852762db506fa3daa155a82bee3c7452adb451adc05e15a
go run ./cmd/app static $previous $current
```
