terraform {
  required_providers {
    chainguard = { source = "chainguard-dev/chainguard" }
    jwt        = { source = "camptocamp/jwt" }
  }
}

variable "group" {
  description = "Group to own the pull token"
  type        = string
}

variable "name" {
  description = "Name of the pull token"
  type        = string
}

variable "ttl_days" {
  description = "TTL of the pull token in days"
  type        = number
  default     = 30
}

resource "tls_private_key" "key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "time_static" "time" {}
resource "random_id" "hex" { byte_length = 4 }

resource "chainguard_identity" "assumable-identity" {
  parent_id = var.group
  name      = var.name

  static {
    expiration = timeadd(time_static.time.rfc3339, "${var.ttl_days * 24}h")
    issuer     = "https://pulltoken.issuer.chainguard.dev"
    subject    = "terraform-pull-token-${random_id.hex.hex}"
    issuer_keys = jsonencode({
      keys = [{
        algorithm = "RS256"
        key       = base64encode(tls_private_key.key.public_key_pem)
      }]
    })
  }
}

// Grant the assumable identity the ability to assume the puller role.
data "chainguard_role" "puller" { name = "registry.pull" }

resource "chainguard_rolebinding" "puller" {
  group    = var.group
  identity = chainguard_identity.assumable-identity.id
  role     = data.chainguard_role.puller.items[0].id
}

resource "jwt_signed_token" "token" {
  algorithm = "RS256"
  claims_json = jsonencode({
    iss = "https://pulltoken.issuer.chainguard.dev"
    iat = time_static.time.unix
    exp = time_static.time.unix + (var.ttl_days * 24 * 60 * 60)
    sub = "terraform-pull-token-${random_id.hex.hex}"
    aud = ["https://issuer.enforce.dev"]
  })
  key = tls_private_key.key.private_key_pem
}

output "username" {
  value = chainguard_identity.assumable-identity.id
}

output "password" {
  value     = jwt_signed_token.token.token
  sensitive = true
}

output "command" {
  value     = <<EOC
docker login -u "${chainguard_identity.assumable-identity.id}" -p "${jwt_signed_token.token.token}" https://cgr.dev
    EOC
  sensitive = true
}
