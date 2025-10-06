#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# Required environment variables
: "${ORG_NAME?ORG_NAME is required.}"
: "${IDENTITY_ID?IDENTITY_ID is required.}"
: "${DST_REPO_NAME?DST_REPO_NAME is required.}"
: "${DST_REPO_URI?DST_REPO_URI is required.}"
: "${UPDATED_WITHIN?UPDATED_WITHIN is required.}"

# Login to AWS ECR
echo "Logging into AWS ECR..." >&2
aws ecr get-login-password \
  | crane auth login --username AWS --password-stdin "${DST_REPO_URI%%/*}"

# Login to Chainguard.
#
# For AWS assumable identities, we have to generate a token by signing a HTTP
# request with AWS credentials and coercing it into a particular format that's
# expected by Chainguard.
#
# Alternatively, you could use a Kubernetes service account token as the
# --identity-token, as described on this page:
# https://edu.chainguard.dev/chainguard/administration/assumable-ids/identity-examples/kubernetes-identity/
echo "Logging into Chainguard..." >&2
eval "$(aws configure export-credentials --format env)"
aws_token=$(curl -X POST "https://sts.amazonaws.com/?Action=GetCallerIdentity&Version=2011-06-15" \
  --aws-sigv4 "aws:amz:us-east-1:sts" \
  --user "${AWS_ACCESS_KEY_ID}:${AWS_SECRET_ACCESS_KEY}" \
  -H "x-amz-security-token: ${AWS_SESSION_TOKEN}" \
  -H "Chainguard-Identity: ${IDENTITY_ID}" \
  -H "Chainguard-Audience: https://issuer.enforce.dev" \
  -H "Accept: application/json" \
  -v 2>&1 > /dev/null \
  | grep '^> ' \
  | sed 's/> //' \
  | base64 -w0
)
chainctl auth login \
  --identity "${IDENTITY_ID}" \
  --identity-token "${aws_token}"
chainctl auth configure-docker \
  --identity "${IDENTITY_ID}" \
  --identity-token "${aws_token}" \
  --audience=cgr.dev

# List every recently updated image.
#
# This produces a list of items with the repo name and tag.
echo "Listing images..." >&2
image_list=$(
  chainctl image list \
    --parent="${ORG_NAME}" \
    --updated-within="${UPDATED_WITHIN}" \
    -o json \
    | jq -cr '.[] | .repo.name as $repo | .tags[] | {repo: $repo, tag: .name}'
)

# If there haven't been any recent updates then the list will be
# empty.
if [[ -z "${image_list}" ]]; then
  echo "No recently updated images found. Exiting." >&2
  exit 0
fi

# Track which repos exist in AWS ECR
declare -A created

# Iterate over each image
echo "Copying images..." >&2
while read -r item; do
  repo=$(jq -r '.repo' <<<"${item}")
  tag=$(jq -r '.tag' <<<"${item}")
  src="cgr.dev/${ORG_NAME}/${repo}:${tag}"
  dst="${DST_REPO_URI}/${repo}:${tag}"

  # Ensure the AWS ECR repository exists 
  if [[ -n created["${repo}"] ]]; then
    if ! aws ecr describe-repositories --repository-names "${DST_REPO_NAME}/${repo}" >/dev/null 2>&1; then
      echo "Creating repository ${DST_REPO_NAME}/${repo}..." >&2
      aws ecr create-repository --repository-name "${DST_REPO_NAME}/${repo}"
    fi
    created["${repo}"]=1
  fi

  # You could use `cosign copy` here if you wanted to also copy the
  # signatures/attestations
  echo "Copying ${src} to ${dst}..." >&2
  crane copy "${src}" "${dst}"
done <<<"${image_list}"
