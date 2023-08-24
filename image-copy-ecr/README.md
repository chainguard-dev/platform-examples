# `image-copy-ecr`

This sets up a Lambda function to listen for `registry.push` events to a private Chainguard Registry group, and mirrors those new images to a repository in Elastic Container Registry.

### Usage

You can use this terraform module to deploy this integration by instantiating
it like this:

```
module "image-copy" {
  source = "github.com/chainguard-dev/enforce-events//image-copy-ecr/iac"

  # The Chainguard IAM group from which we expect to receive events.
  # This is used to authenticate that the Chainguard events are intended
  # for you, and not another user.
  # Images pushed to repos under this group will be mirrored to Artifact Registry.
  group = "<group-id>"

  # This is the location in ECR where images will be mirrored.
  # For example: pushes to cgr.dev/<group>/foo:1.2.3 will be mirrored to
  # <account>.dkr.ecr.<region>.amazonaws.com/<dst_repo>/foo:1.2.3
  dst_repo = "mirrored/images"
}
```

The Terraform does everything:

- builds the mirroring app into an image using `ko_build`
- deploys the app to a Lambda function
- sets up a Chainguard Identity with permissions to pull from the private cgr.dev repo
- allows the Lambda function to assume the puller identity and push to ECR
- sets up a subscription to notify the Lambda function when pushes happen to cgr.dev

### Setup

```sh
aws sso login --profile my-profile
chainctl auth login
terraform init
terraform apply
```

When the resources are created, any images that are pushed to your group will be mirrored to the ECR repository.

The Lambda function has minimal permissions: it's only allowed to push images to the destination repo and its sub-repos.

The Chainguard identity also has minimal permissions: it only has permission to pull from the source repo.

To tear down resources, run `terraform destroy`.
