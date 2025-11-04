# `image-copy-ecr-cronjob`

Runs a regular `CronJob` in an AWS EKS cluster that copies recently updated
images from a Chainguard organization to AWS ECR.

## Requirements

1. You must have an existing AWS EKS cluster.
2. And the following tools available locally:
   - `aws`
   - `chainctl`
   - `docker`
   - `kubectl`
   - `terraform`

## Usage

You can use the provided Terraform module to quickly create the required
resources:

- An AWS ECR repository to host the image for the `image-copy` script.
- An AWS ECR repository to host your copied Chainguard images
- An AWS role that grants the `image-copy` job permissions for AWS ECR
- An assumable identity that allows the job to list images via the Chainguard
  APIs
- A Kubernetes namespace (`chainguard`) and a cron job (`image-copy`)

Follow these steps.

1. Login to Chainguard, AWS, AWS ECR and configure your Kubernetes context.

```
chainctl auth login
chainctl auth configure-docker
aws sso login
aws ecr get-login-password --region <region> | docker login --username AWS --password-stdin <account-id>.dkr.ecr.<region>.amazonaws.com
aws eks update-kubeconfig --region=<region> --name=<cluster-name>
```

2. Apply the Terraform code.

```
cd iac/

cat <<EOF > terraform.tfvars
# Required. The name of your Chainguard organization.
org_name = "your.org"

# Required. The name of your AWS EKS cluster. The cluster must already exist.
cluster_name = "your-cluster-name"

# Required. The name of the AWS ECR repository to copy images to
repo_name = "chainguard"

# Optional. The Kubernetes namespace for the CronJob. Terraform will create
# this namespace. Defaults to 'chainguard'.
namespace = "chainguard"

# Optional. The name of the Chainguard image to build the image-copy image
# from. Must have apk-tools and a shell. Defaults to chainguard-base.
base_image_name = "chainguard-base:latest"

# Optional. The platform of the image to build. This should match the platform
# of the nodes in your AWS EKS cluster.
image_platform = "linux/amd64"

# Optional. The job will only copy images that have been updated within this
# period of time. Defaults to 72h.
updated_within = "72h"
EOF

terraform init

terraform apply -var-file terraform.tfvars
```

You can trigger the copy ahead of schedule by creating a job.

```
kubectl create job image-copy -n chainguard --from=cronjob/image-copy
kubectl logs -n chainguard job/image-copy
```

## Usage (The Hard Way)

Here's how to do everything the Terraform does yourself with CLI commands.

1. Export some variables for use in subsequent steps.

```
export CLUSTER_NAME="<your-cluster-name>"
export AWS_ACCOUNT_ID="<your-account-id>"
export AWS_REGION="<your-region>"
exporg ORG_NAME="<your-chainguar-org-name>"
```

2. Login to Chainguard, AWS, AWS ECR and configure your kubeconfig.

```
chainctl auth login
chainctl auth configure-docker
aws sso login
aws ecr get-login-password --region "${AWS_REGION}" \
    | docker login --username AWS --password-stdin "${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
aws eks update-kubeconfig --region="${AWS_REGION}" --name="${CLUSTER_NAME}"
```

3. Create a parent repository for your Chainguard images.

```
aws ecr create-repository --repository-name chainguard
```

4. Create a repository for the `image-copy` image.

```
aws ecr create-repository --repository-name chainguard-image-copy
```

5. Build and push the `image-copy` image. Substitute in your organization name,
   AWS account ID and region. You can replace `BASE_IMAGE_NAME` if you don't
   have access to `chainguard-base` in your organization. It must be an image
   with a shell and `apk-tools`.

```
docker build . --push \
    --build-arg ORG_NAME=${ORG_NAME} \
    --build-arg BASE_IMAGE_NAME=chainguard-base:latest \
    --provenance=false \
    -t "${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/chainguard-image-copy"
```

6. Create an AWS role with the required permissions.

```
# Get the details of the cluster OIDC provider
OIDC_PROVIDER=$(aws eks describe-cluster --name $CLUSTER_NAME --query "cluster.identity.oidc.issuer" --output text | sed 's|https://||')
OIDC_PROVIDER_ARN=$(aws iam list-open-id-connect-providers --query "OpenIDConnectProviderList[?ends_with(Arn, '${OIDC_PROVIDER##*/}')].Arn" --output text)

# Create the trust policy that allows the chainguard/image-copy service account
# to assume the role
cat > trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "$OIDC_PROVIDER_ARN"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "${OIDC_PROVIDER}:sub": "system:serviceaccount:chainguard:image-copy",
          "${OIDC_PROVIDER}:aud": "sts.amazonaws.com"
        }
      }
    }
  ]
}
EOF

# Create the IAM role
aws iam create-role \
  --role-name "${CLUSTER_NAME}-image-copy" \
  --assume-role-policy-document file://trust-policy.json

# Get the arn of the AWS ECR repository
ECR_ARN=$(aws ecr describe-repositories --repository-names chainguard --query "repositories[0].repositoryArn --output text)

# Create the policy that allows the image-copy role to copy images to the AWS
# ECR repository
cat <<EOF> inline-policy.json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:CreateRepository",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:GetRepositoryPolicy",
        "ecr:DescribeRepositories",
        "ecr:ListImages",
        "ecr:DescribeImages",
        "ecr:BatchGetImage",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload",
        "ecr:PutImage"
      ],
      "Resource": [
        "$ECR_ARN",
        "${ECR_ARN}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken"
      ],
      "Resource": "*"
    }
  ]
}
EOF

# Attach the policy to the role
aws iam put-role-policy \
  --role-name "${CLUSTER_NAME}-image-copy" \
  --policy-name "image-copy" \
  --policy-document file://inline-policy.json
```

7. Create a Chainguard identity which can be assumed by the AWS role. Save the
   id to a variable for later use.

```
cat > id.json <<EOF
{
   "name": "${CLUSTER_NAME}-image-copy",
   "awsIdentity": {
     "aws_account" : "${AWS_ACCOUNT_ID}",
     "userIdPattern"  : "^AROA(.*):(.*)$",
     "arnPattern" : "arn:aws:sts::${AWS_ACCOUNT_ID}:assumed-role/${CLUSTER_NAME}-image-copy/(.*)"
   }
}
EOF

export IDENTITY_ID=$(chainctl iam id create aws --parent "${ORG_NAME}" --filename id.json -o id)
```

8. Deploy the CronJob to the AWS EKS cluster.

```
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: chainguard
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: image-copy
  namespace: chainguard
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::${AWS_ACCOUNT_ID}:role/${CLUSTER_NAME}-image-copy
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: image-copy
  namespace: chainguard
spec:
  concurrencyPolicy: Replace
  failedJobsHistoryLimit: 5
  schedule: 1 0 * * *
  successfulJobsHistoryLimit: 3
  suspend: false
  jobTemplate:
    spec:
      backoffLimit: 2
      completions: 1
      parallelism: 1
      template:
        spec:
          serviceAccountName: image-copy
          restartPolicy: Never
          containers:
          - name: image-copy
            image: ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/chainguard-image-copy
            imagePullPolicy: Always
            env:
            - name: ORG_NAME
              value: ${ORG_NAME}
            - name: IDENTITY_ID
              value: ${IDENTITY_ID}
            - name: DST_REPO_NAME
              value: chainguard
            - name: DST_REPO_URI
              value: ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/chainguard
            - name: UPDATED_WITHIN
              value: 72h
            - name: AWS_REGION
              value: ${AWS_REGION} 
EOF
```

9. Trigger the copy ahead of schedule by creating a job.

```
kubectl create job image-copy -n chainguard --from=cronjob/image-copy
kubectl logs -n chainguard job/image-copy
```
