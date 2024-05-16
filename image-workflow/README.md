# `image-workflow`

This sets up a Cloud Run app to listen for `registry.push` events to a private Chainguard Registry group, and triggers a GitHub Actions workflow with that image ref as an input.

This lets you define GitHub Actions workflows to pull and test images in response to pushes.

# Usage

To start, create a GitHub workflow at `.github/workflows/workflow.yaml`, with an input named `image`:

```
on:
  workflow_dispatch:
    inputs:
      image:
        description: 'Image to test'
        required: true

jobs:
  test-image:
    runs-on: ubuntu-latest
    steps:
      # This sets up auth with the registry to be able to pull the image.
      - uses: chainguard-dev/actions/setup-chainctl@main
        with:
          identity: TODO  # We'll fill this in later.

      - run: |
          # Your tests go here.
          docker pull ${{ github.event.inputs.image }}
```

Then `terraform apply` the module (e.g., from the root of this repo):

```
module "image-workflow" {
  source = "./image-workflow/iac"  # TODO: move to enforce-events

  # name is used to prefix resources created by this demo application
  # where possible.
  name = "chainguard-dev"

  # This is the GCP project ID in which certain resource will live including:
  #  - The container image for this application,
  #  - The Cloud Run service hosting this application,
  project_id = "<project-id>"

  # The Chainguard IAM group from which we expect to receive events.
  # This is used to authenticate that the Chainguard events are intended
  # for you, and not another user.
  # Images pushed to repos under this group will trigger workflows.
  group = "<group-id>"

  # These describe the GitHub organization, repository and workflow to trigger.
  github_org         = "my-org"
  github_repo        = "my-repo"
  github_workflow_id = "workflow.yaml"

  # Location of the Cloud Run subscriber.
  # location = "us-central1" (default)
}
```

## Setting your GitHub Personal Access Token

Once things have been provisioned, this module outputs a `secret-command`
containing the command to run to upload your Github "personal access token" to
the Google Secret Manager secret the application will use, looking something
like this:

```shell
echo -n YOUR GITHUB PAT | \
  gcloud --project ... secrets versions add ... --data-file=-
```

The personal access token needs `actions:write` to trigger workflows.

## Setting your puller identity

The module also outputs the identity that was created, which can be assumed by the
GitHub Actions workflow. This is the value that goes in the `setup-chainctl` step
of your workflow above.
