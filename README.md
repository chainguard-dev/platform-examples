# Chainguard Platform Examples

This repo holds a number of example apps demonstrating integrations with Chainguard platform APIs, and various [Chainguard Events](https://edu.chainguard.dev/chainguard/chainguard-enforce/reference/events/).

- [GitHub Issue Opener](./github-issue-opener/README.md) - opens an issue in GitHub when a policy is violated
- [Slack Webhook](./slack-webhook/README.md) - sends a message to a Slack channel when a policy is violated
- [Jira Issuer Opener](./jira-issue-opener/) - opens an issue in Jira when a policy is violated
- [GCP Image Copier](./image-copy-gcp/) - copies images to Google Artifact Registry when an image is pushed to cgr.dev
- [ECR Image Copier](./image-copy-ecr/) - copies images to Amazon Elastic Container Registry when an image is pushed to cgr.dev
- [AWS Auth Example](./aws-auth/) - demonstrates configuration of an AWS assumable Chainguard identity, as well as calling the Chainguard API from a Lambda function
- [Tag History Example](./tag-history/) - demonstrates how to use the Chainguard API to track tag history for images in a registry
- [Image Diff Example](./image-diff/) - demonstrates how to use the Chainguard API to compare images in a registry

> [!NOTE]
> These examples are intended to be used as a reference for building your own Chainguard platform integrations.
> They can be used directly as-is, but are not intended to be production-ready and may experience breaking changes or be removed entirely.
> You can reference these examples in your own Terraform configs, but we recommend that you pin a specific commit to avoid unexpected changes.
> For example:

```hcl
module "github-issue-opener" {
  source = "github.com/chainguard-dev/platform-examples//github-issue-opener/iac?ref=a1b2c3d4"

  project_id = "..."
  group      = "..."
}
