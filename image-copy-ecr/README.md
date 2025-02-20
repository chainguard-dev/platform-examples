# `image-copy-ecr`

This sets up a Lambda function to listen for `registry.push` events to a private Chainguard Registry group, and mirrors those new images to a repository in Elastic Container Registry.

The Terraform does everything:

- builds the mirroring app into an image using `ko_build`
- deploys the app to a Lambda function
- sets up a Chainguard Identity with permissions to pull from the private cgr.dev repo
- allows the Lambda function to assume the puller identity and push to ECR
- sets up a subscription to notify the Lambda function when pushes happen to cgr.dev

## Setup

```sh
aws sso login --profile my-profile
chainctl auth login
terraform init
terraform apply
```

This will prompt for your group name and destination repo, and show you the resources it will create.

When the resources are created, any images that are pushed to your group will be mirrored to the ECR repository.

The Lambda function has minimal permissions: it's only allowed to push images to the destination repo and its sub-repos.

The Chainguard identity also has minimal permissions: it only has permission to pull from the source repo.

To tear down resources, run `terraform destroy`.
