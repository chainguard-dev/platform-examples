#!/usr/bin/env bash
set -euo pipefail
trap "echo 'Interrupted. Exiting...'; exit 1" SIGINT SIGTERM

AWS_PAGER=""
AWS_REGION="ap-southeast-2"
TARGET_ECR_REPO_PREFIX="chainguard" # Change this to your desired ECR repo prefix
TARGET_ECR="452336408843.dkr.ecr.ap-southeast-2.amazonaws.com"
CG_ORG="andrewd.dev"

chainctl images repos list --parent="$CG_ORG" -o json \
  | jq -r '.items[].name' \
  | while read -r image; do
    echo "Creating ECR repo: $TARGET_ECR$TARGET_ECR_REPO_PREFIX/$image"
    aws ecr create-repository \
      --region $AWS_REGION \
      --repository-name "$TARGET_ECR_REPO_PREFIX/$image" \
      --output text \
      --query 'repository.repositoryUri' \
      2>/dev/null || echo "ECR repo $image may already exist"

    chainctl img ls --parent="$CG_ORG" --repo "$image" -o json \
      | jq -r '.[].tags[].name' \
      | while read -r tag; do
          echo "#### Copying: cgr.dev/$CG_ORG/${image}:${tag} -> $TARGET_ECR/$TARGET_ECR_REPO_PREFIX/${image}:${tag} ####"
          crane cp "cgr.dev/$CG_ORG/${image}:${tag}" "$TARGET_ECR/$TARGET_ECR_REPO_PREFIX/${image}:${tag}"
        done
  done