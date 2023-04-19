module "slack-webhook" {
  source     = "github.com/chainguard-dev/enforce-events/slack-webhook/iac"
  name       = "enforce-events"
  project_id = "jamon-chainguard"
  group      = "jamon-test"
}
