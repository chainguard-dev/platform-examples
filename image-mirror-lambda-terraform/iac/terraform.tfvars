aws_region  = "us-east-2"
aws_profile = "cg-dev"

group_name  = "bannon.dev"
name_prefix = "image-copy-all"

# optional: dst_prefix = "mirrors"

# identity id (username) for your pull token
cgr_username = "<PULL_TOKEN_USERNAME>"

mirror_dry_run = false

repo_list = [
  "cgr.dev/bannon.dev/datadog-agent",
  "cgr.dev/bannon.dev/node",
  "cgr.dev/bannon.dev/python",
  "cgr.dev/bannon.dev/jdk",
  "cgr.dev/bannon.dev/jre",
  "cgr.dev/bannon.dev/envoy",
]

repo_tags = {
  "cgr.dev/bannon.dev/node"        = ["22"]
  "cgr.dev/bannon.dev/datadog-agent" = ["7.69", "7.69-dev"]
  "cgr.dev/bannon.dev/python" =    ["3.11", "3.11-dev"]
  "cgr.dev/bannon.dev/jdk" =   ["openjdk-11", "openjdk-17"]
  "cgr.dev/bannon.dev/jre" =   ["openjdk-11", "openjdk-17"]
}

copy_all_tags = true
