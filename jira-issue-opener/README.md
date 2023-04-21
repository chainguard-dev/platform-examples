## Jira Issue Opener

This demo application shows how users can write a very simple application that
authenticates Chainguard webhook requests for continuous verification policy
violations, and turns them into Jira issues.

### Usage

You can use this terraform module to deploy this integration by instantiating
it like this:

```hcl
# TODO: pre-reqs like ko/google providers.

module "jira-issue-opener" {
  source = "github.com/chainguard-dev/enforce-events/jira-issue-opener/iac"

  # name is used to prefix resources created by this demo application
  # where possible.
  name = "chainguard-dev"

  # This is the GCP project ID in which certain resource will live including:
  #  - The container image for this application,
  #  - The Cloud Run service hosting this application,
  #  - The Secret Manager secret holding the github access token
  #    for opening issues.
  project_id = var.gcp_project_id

  # The Chainguard IAM group from which we expect to receive events.
  # This is used to authenticate that the Chainguard events are intended
  # for you, and not another user.
  group = var.chainguard_iam_group

  # These describe the Jira user and project in which issues
  # will be opened.
  jira_user = "someone@chainguard.dev"
  jira_project = "EE"

  # The issue type to create e.g. task, bug, or a custom type
  issue_type = "task"
}
```

Once things have been provisioned, this module outputs a `secret-command`
containing the command to run to upload your Atlassian API token to the
Google Secret Manager secret the application will use, looking something
like this:

```shell
echo -n YOUR ATLASSIAN API TOKEN | \
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
Jira issues outlining the policy violation.
