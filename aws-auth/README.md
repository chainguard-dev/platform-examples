# `aws-auth`

This sets up a Lambda function that executes with permission to list Chainguard repositories.
This demonstrates configuration of an AWS assumable Chainguard identity, as well as calling the Chainguard API from a Lambda function.

The Terraform does everything:

- builds the example app into an image using `ko_build`
- deploys the app to a Lambda function
- sets up a Chainguard Identity with permissions to list cgr.dev repos
- allows the Lambda function to assume the Chainguard identity
- hosts a public URL that can be invoked to list repos

## Setup

```sh
aws sso login --profile my-profile
terraform init
terraform apply
```

This will prompt for a group ID, and show you the resources it will create, as well as a public URL you can visit.

When the resources are created, visiting the URL will list repos in the specified group.

The Lambda function has no AWS permissions, and can only view Chainguard images.

To tear down resources, run `terraform destroy`.
