# `terraform-pull-token`

This module creates a token that can be used to pull images from a private registry.
It's intended to be equivalent to [`chainctl auth configure-docker --pull-token`](https://edu.chainguard.dev/chainguard/chainctl/chainctl-docs/chainctl_auth_configure-docker/).

### Usage

```
terraform apply \
  -var=group=[ROOT-GROUP-ID] \
  -var=name=[UNIQUE-IDENTITY-NAME]
```

When this completes it will output some values:

```
Outputs:

command = <sensitive>
password = <sensitive>
username = "7bf0abc1e47a927f78fabc9cd6393f4a83ec6eb7/2d3d8f583b8df855"
```

The `password` in sensitive, which means you need to use the `terraform output -raw` command to see it.

```
$ terraform output -raw command
docker login -u "7bf0abc1e47a927f78fabc9cd6393f4a83ec6eb7/2d3d8f583b8df855" -p "eyJhbG...oMtQ" https://cgr.dev
```

You can run this command to log in using the pull token you just created.

