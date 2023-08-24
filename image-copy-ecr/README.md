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

This will prompt for a group ID and destination repo, and show you the resources it will create.

When the resources are created, any images that are pushed to your group will be mirrored to the ECR repository.

The Lambda function has minimal permissions: it's only allowed to push images to the destination repo and its sub-repos.

The Chainguard identity also has minimal permissions: it only has permission to pull from the source repo.

To tear down resources, run `terraform destroy`.

## Demo

After setting up the infrastructure as described above:

```sh
crane cp random.kontain.me/random cgr.dev/<org>/random:hello-demo
```

This pulls a randomly generated image from `kontain.me` and pushes it to your private registry.

The Lambda function you set up will fire and copy the image to ECR. A few seconds later:

```sh
crane ls <account-id>.dkr.ecr.<region>.amazonaws.com/<dst-repo>/random
hello-demo
```

It worked! 🎉
