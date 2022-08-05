## Slack Webhook

This demo application shows how users can write a very simple application that
authenticates Chainguard webhook requests for continuous verification policy
violations, and turns them into Slack notifications.

### Usage

You can use this terraform module to deploy this integration by instantiating
it like this:

```hcl
# TODO: pre-reqs like ko/google providers.

module "issue-opener" {
  # TODO: Replace with whatever we name this.
  source = "github.com/chainguard-dev/sample-slack-notifier//iac"

  # name is used to prefix resources created by this demo application
  # where possible.
  name = "chainguard-dev"

  # This is the GCP project ID in which certain resource will live including:
  #  - The container image for this application,
  #  - The Cloud Run service hosting this application,
  #  - The Secret Manager secret holding the slack webhook URL
  #    for opening issues.
  project_id = var.gcp_project_id

  # The Chainguard environment that will be sending us events.
  # This is used to authenticate the events we receive are from Chainguard.
  env = "enforce.dev"

  # The Chainguard IAM group from which we expect to receive events.
  # This is used to authenticate that the Chainguard events are intended
  # for you, and not another user.
  group = var.chainguard_iam_group

  # This let's you control the verbosity of the notifications.
  # WARN  - policy failures
  # INFO  - policy failures or improvements
  # DEBUG - all events
  notify_level = var.notification_level
}
```

Once things have been provisioned, this module outputs a `secret-command`
containing the command to run to upload your Slack webhook URL to the Google
Secret Manager secret the application will use, looking something like this:

```shell
echo -n YOUR GITHUB PAT | \
  gcloud --project ... secrets versions add ... --data-file=-
```

Create a Slack Webhook URL as detailed in the
[Incoming webhooks for Slack](https://slack.com/help/articles/115005265063-Incoming-webhooks-for-Slack)
help document, and run the above command!


Once the secret has been setup, grab the `url` output and the `group` we passed
in and run:

```shell
chainctl event subscriptions create URL --group=GROUP
```

That's it!  Now policy failures during continuous verification will
post notifications to slack outlining the policy violation.
