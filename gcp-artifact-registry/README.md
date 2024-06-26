# Configuring Google Artifact Registry Pull-Through Cache for cgr.dev

This directory contains Terraform resources to configure a [Google Artifact Registry remote repository](https://cloud.google.com/artifact-registry/docs/repositories/remote-repo) for a private `cgr.dev` registry, using a [Pull Token](https://edu.chainguard.dev/chainguard/chainguard-registry/authenticating/#authenticating-with-a-pull-token).

First, create a pull token for your organization:

```
chainctl auth configure-docker --pull-token
```

This will print a `docker login` command that you can use to authenticate with the pull token. For example:

```
To use this pull token in another environment, run this command:

    docker login "cgr.dev" --username "abcdef11e47a927f78ffa39cd6393f4a83ec6eb7/e3b0123abcdef123" --password "eyJh...krglE"
```

We'll use this username and password to configure the pull-through cache in Terraform.

Next, `terraform init` and `terraform apply` the configuration in this directory:

```
terraform apply
```

This will prompt for your GCP project, and the Chainguard pull token `username` and `password`.

When Terraform is done creating resources, it will print the remote repo URLs that were created.

```
repos = [
  "us-east4-docker.pkg.dev/my-cool-project/remote",
]
```

You can now pull through these URLs to pull images from the `cgr.dev` registry.

Simply append your Chainguard organization name and repo name to the URL, and use it as a remote repo URL.

```
docker pull us-east4-docker.pkg.dev/my-cool-project/remote/customer.biz/image:latest
latest: Pulling from my-cool-project/remote/customer.biz/image
6cdb74e61124: Downloading [====>                                              ]  18.21MB/224.2MB
```

You can configure Terraform to set up resources in other regions, or with other names, by updating these variables:

```
TF_VAR_regions="['us-central1']" TF_VAR_repo="custom" terraform apply
...
repos = [
  "us-central1-docker.pkg.dev/my-cool-project/custom",
]
```
