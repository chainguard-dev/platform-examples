# Image Mirror Script

A bash script that mirrors container images from Chainguard registry to AWS ECR using `chainctl` and `crane`.

## Overview

This script automates the process of:
1. Listing all image repositories in a specified Chainguard organization
2. Creating corresponding ECR repositories in AWS
3. Copying all tagged images from Chainguard registry to ECR

## Prerequisites

- `chainctl` - Chainguard CLI tool
- `aws` - AWS CLI configured with appropriate permissions
- `crane` - Container image manipulation tool
- `jq` - JSON processor

## Configuration

Edit the following variables in the script before running:

```bash
AWS_REGION="ap-southeast-2"                          # Your AWS region
TARGET_ECR_REPO_PREFIX="/chainguard"                 # ECR repo prefix (or leave empty)
TARGET_ECR="452336408843.dkr.ecr.ap-southeast-2.amazonaws.com"  # Your ECR registry URL
CG_ORG="andrewd.dev"                                 # Your Chainguard organization
```

## Required AWS Permissions

Your AWS credentials need the following ECR permissions:
- `ecr:CreateRepository`
- `ecr:BatchCheckLayerAvailability`
- `ecr:InitiateLayerUpload`
- `ecr:UploadLayerPart`
- `ecr:CompleteLayerUpload`
- `ecr:PutImage`

## Usage

1. Ensure you're authenticated with both Chainguard and AWS:
   ```bash
   chainctl auth login
   aws configure
   ```

2. Make the script executable:
   ```bash
   chmod +x mirror-images.sh
   ```

3. Run the script:
   ```bash
   ./mirror-images.sh
   ```

## What the Script Does

1. **List Repositories**: Uses `chainctl` to get all image repositories in the specified organization
2. **Create ECR Repos**: For each repository, creates a corresponding ECR repository (skips if already exists)
3. **Mirror Images**: Lists all tags for each repository and copies them using `crane`

## Output

The script provides detailed logging showing:
- ECR repository creation status
- Each image being copied with source and destination URLs

## Safety Features

- `set -euo pipefail` ensures the script exits on any error
- Signal handling for clean interruption (Ctrl+C)
- Error suppression for ECR repo creation (continues if repo already exists)

## Example Output

```
Creating ECR repo: /chainguard/my-app
#### Copying: cgr.dev/andrewd.dev/my-app:latest -> 452336408843.dkr.ecr.ap-southeast-2.amazonaws.com/chainguard/my-app:latest ####
#### Copying: cgr.dev/andrewd.dev/my-app:v1.0.0 -> 452336408843.dkr.ecr.ap-southeast-2.amazonaws.com/chainguard/my-app:v1.0.0 ####
```