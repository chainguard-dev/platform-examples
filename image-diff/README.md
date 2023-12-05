# `image-diff`

This demonstrates how to query diffs of Chainguard Images.

Along with Tag History, you can use this to show the evolution of an image over time.

### Usage


```sh
group=customer.biz # defaults to public images
reponame=kubeflow-pipelines-frontend
previous=sha256:00d89a080976cad0fb94e1c1629c9f3a4d8d424572ba590955147348c80a0473
current=sha256:54382392baa1e116f3540c6b1c96c7ae441d06b38504e8be1546b85c9405e94b
go run ./cmd/app --group=$group $reponame $previous $current
```
