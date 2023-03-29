## Github Issue Opener

This demo application shows how users can write a very simple application that
authenticates Chainguard webhook requests for continuous verification policy
violations, and turns them into Github issues.

### Usage

You can use this terraform module to deploy this integration by instantiating
it like this:

```hcl
# TODO: pre-reqs like ko/google providers.

module "issue-opener" {
  source = "github.com/chainguard-dev/enforce-events//github-issue-opener/iac"

  # name is used to prefix resources created by this demo application
  # where possible.
  name = "chainguard-dev"

  # This is the GCP project ID in which certain resource will live including:
  #  - The container image for this application,
  #  - The Cloud Run service hosting this application,
  #  - The Secret Manager secret holding the github access token
  #    for opening issues.
  project_id = var.gcp_project_id

  # The Chainguard environment that will be sending us events.
  # This is used to authenticate the events we receive are from Chainguard.
  env = "enforce.dev"

  # The Chainguard IAM group from which we expect to receive events.
  # This is used to authenticate that the Chainguard events are intended
  # for you, and not another user.
  group = var.chainguard_iam_group

  # These describe the github organization and repository in which github issues
  # will be opened.
  github_org  = "chainguard-dev"
  github_repo = "mono"

  # These are the labels that get applied to opened issues.
  labels = "label1,label2,label3"
}
```

Once things have been provisioned, this module outputs a `secret-command`
containing the command to run to upload your Github "personal access token" to
the Google Secret Manager secret the application will use, looking something
like this:

```shell
echo -n YOUR GITHUB PAT | \
  gcloud --project ... secrets versions add ... --data-file=-
```

The personal access token needs permission to open issues on the target
repository.


Once the secret has been setup, grab the `url` output and the `group` we passed
in and run:

```shell
chainctl event subscriptions create URL --group=GROUP
```

That's it!  Now policy failures during continuous verification will open
github issues outlining the policy violation.
