# `image-copy-ecr`

This sets up a Lambda function to listen for `registry.push` events to a private Chainguard Registry group, and mirrors those new images to a repository in Elastic Container Registry.

The Terraform does everything:

- builds the mirroring app into an image using `ko_build`
- deploys the app to a Lambda function
- sets up a Chainguard Identity with permissions to pull from the private cgr.dev repo
- allows the Lambda function to assume the puller identity and push to ECR
- sets up a subscription to notify the Lambda function when pushes happen to cgr.dev

## Setup

Create a `.tfvars` file.

```
cat << EOF > iac/terraform.tfvars
# Required. The name of your Chainguard organization.
group_name = "your.org"

# Required. The name of the destination repo where images should be copied to.
# This repository will be created by the terraform and images will be copied to
# '<dst_repo>/<image_name>'.
dst_repo = "image-copy"

# Optional. Ignore signatures and attestations. This can help reduce cruft in
# the mirror repositories if you aren't going to be verifying or using the
# referrers.
# ignore_referrers = true

# Optional. Enable immutable tags for the repositories created by the Lambda.
# If enabled, then the Lambda will append a portion of the digest to the tags
# it copies. For instance: 'latest-abcdef'
# immutable_tags = true
EOF
```

Login to AWS and Chainguard.

```sh
aws sso login --profile my-profile
chainctl auth login
```

Apply the terraform.

```sh
cd iac/
terraform init
terraform apply -var-file=terraform.tfvars
```

When the resources are created, any images that are pushed to your group will
be mirrored to the ECR repository.

The Lambda function has minimal permissions: it's only allowed to push images
to the destination repo and its sub-repos.

The Chainguard identity also has minimal permissions: it only has permission to
pull from the source repo.

To tear down resources, run `terraform destroy -var-file=terraform.tfvars`.
