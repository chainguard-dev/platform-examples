# `image-copy-gcp`

This sets up a Cloud Run app to listen for [`registry.push` events](https://edu.chainguard.dev/chainguard/chainguard-enforce/reference/events/#service-registry---push) to a private Chainguard Registry group, and mirrors those new images to a repository in Google Artifact Registry.

### Usage

You can use this terraform module to deploy this integration by instantiating
it like this:

```
module "image-copy" {
  source = "github.com/chainguard-dev/platform-examples//image-copy-gcp/iac"

  # name is used to prefix resources created by this demo application
  # where possible.
  name = "chainguard-dev"

  # This is the GCP project ID in which certain resource will live including:
  #  - The container image for this application, and mirrored images,
  #  - The Cloud Run service hosting this application,
  #  - The Service Account that authorizes pushes to Google Artifact Registry.
  project_id = "<project-id>"

  # The name of the Chainguard IAM group from which we expect to receive events.
  # This is used to authenticate that the Chainguard events are intended
  # for you, and not another user.
  # Images pushed to repos under this group will be mirrored to Artifact Registry.
  group_name = "<group-name>"

  # This is the location in Artifact Registry where images will be mirrored.
  # For example: pushes to cgr.dev/<group>/foo will be mirrored to
  # <location>-docker.pkg.dev/<project_id>/<dst_repo>/foo.
  dst_repo = "mirrored/images"

  # Location of the Artifact Registry repository, and the Cloud Run subscriber.
  # location = "us-central1" (default)

  # Don't copy referrers like signatures and attestations
  # ignore_referrers = true

  # Verify signatures before copying an image
  # verify_signatures = true
}
```

To use it, `chainctl auth login` and `terraform apply`.

The Terraform does everything:

- builds the mirroring app into an image using `ko_build`
- deploys the app to a Cloud Run service, with permission to push to Google Artifact Registry
- sets up a Chainguard Identity with permissions to pull from the private cgr.dev repo
- allows the Cloud Run service's SA to assume the puller identity
- sets up a subscription to notify the Cloud Run service when pushes happen to cgr.dev
