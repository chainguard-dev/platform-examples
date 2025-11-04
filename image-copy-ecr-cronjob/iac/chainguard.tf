data "aws_caller_identity" "current" {}

data "chainguard_group" "org" {
  name = var.org_name
}

resource "chainguard_identity" "image_copy" {
  parent_id   = data.chainguard_group.org.id
  name        = "image-copy"
  description = "Identity for image-copy job."

  aws_identity {
    aws_account         = data.aws_caller_identity.current.account_id
    aws_user_id_pattern = "^AROA(.*):(.*)$"
    aws_arn_pattern     = "^arn:aws:sts::${data.aws_caller_identity.current.account_id}:assumed-role/${aws_iam_role.image_copy.name}/(.*)$"

  }
}

# Look up the registry.pull role to grant the identity.
data "chainguard_role" "puller" {
  name = "registry.pull"
}

resource "chainguard_rolebinding" "puller" {
  identity = chainguard_identity.image_copy.id
  role     = data.chainguard_role.puller.items[0].id
  group    = data.chainguard_group.org.id
}
