package mapper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Repo describes a repo in the catalog
type Repo struct {
	Name        string   `json:"name"`
	CatalogTier string   `json:"catalogTier"`
	Aliases     []string `json:"aliases"`
}

func listRepos(ctx context.Context) ([]Repo, error) {
	c := &http.Client{}

	buf := bytes.NewReader([]byte(`{"query":"query OrganizationImageCatalog($organization: ID!) {\n  repos(filter: {uidp: {childrenOf: $organization}}) {\n    name\n    aliases\n  catalogTier\n  }\n}","variables":{"excludeDates":true,"excludeEpochs":true,"organization":"ce2d1984a010471142503340d670612d63ffb9f6"}}`))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://data.chainguard.dev/query?id=PrivateImageCatalog", buf)
	if err != nil {
		return nil, fmt.Errorf("constructing request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "image-mapper")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data struct {
		Data struct {
			Repos []Repo `json:"repos"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("unmarshaling body: %w", err)
	}

	return fixAliases(data.Data.Repos), nil
}

// fixAliases corrects some notoriously incorrect aliases in the repository
// data. Generally these are cases where we associate multiple images in the
// same 'family' with every image in the 'family'.
//
// Naturally, this should be fixed in the actual data but that's
// non-trivial to do at the moment. So, until such time, we'll do it here to
// improve the results in the short term.
func fixAliases(repos []Repo) []Repo {
	for i, repo := range repos {
		for name, aliases := range aliasesFixes {
			if repo.Name != name {
				continue
			}
			repos[i].Aliases = aliases
		}
	}

	return repos
}

var aliasesFixes = map[string][]string{
	"argo-cli": {
		"quay.io/argoproj/argocli",
	},
	"argo-cli-fips": {
		"quay.io/argoproj/argocli",
	},
	"argo-events": {
		"quay.io/argoproj/argo-events",
	},
	"argo-events-fips": {
		"quay.io/argoproj/argo-events",
	},
	"argo-exec": {
		"quay.io/argoproj/argoexec",
	},
	"argo-exec-fips": {
		"quay.io/argoproj/argoexec",
	},
	"argo-workflowcontroller": {
		"quay.io/argoproj/workflow-controller",
	},
	"argo-workflowcontroller-fips": {
		"quay.io/argoproj/workflow-controller",
	},
	"crossplane-aws": {
		"ghcr.io/crossplane-contrib/provider-family-aws",
	},
	"crossplane-aws-cloudformation": {
		"ghcr.io/crossplane-contrib/provider-aws-cloudformation",
	},
	"crossplane-aws-cloudformation-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-cloudformation",
	},
	"crossplane-aws-cloudfront": {
		"ghcr.io/crossplane-contrib/provider-aws-cloudfront",
	},
	"crossplane-aws-cloudfront-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-cloudfront",
	},
	"crossplane-aws-cloudwatchlogs": {
		"ghcr.io/crossplane-contrib/provider-aws-cloudwatchlogs",
	},
	"crossplane-aws-cloudwatchlogs-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-cloudwatchlogs",
	},
	"crossplane-aws-dynamodb": {
		"ghcr.io/crossplane-contrib/provider-aws-dynamodb",
	},
	"crossplane-aws-dynamodb-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-dynamodb",
	},
	"crossplane-aws-ec2": {
		"ghcr.io/crossplane-contrib/provider-aws-ec2",
	},
	"crossplane-aws-ec2-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-ec2",
	},
	"crossplane-aws-eks": {
		"ghcr.io/crossplane-contrib/provider-aws-eks",
	},
	"crossplane-aws-eks-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-eks",
	},
	"crossplane-aws-fips": {
		"ghcr.io/crossplane-contrib/provider-family-aws",
	},
	"crossplane-aws-firehose": {
		"ghcr.io/crossplane-contrib/provider-aws-firehose",
	},
	"crossplane-aws-firehose-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-firehose",
	},
	"crossplane-aws-iam": {
		"ghcr.io/crossplane-contrib/provider-aws-iam",
	},
	"crossplane-aws-iam-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-iam",
	},
	"crossplane-aws-kinesis": {
		"ghcr.io/crossplane-contrib/provider-aws-kinesis",
	},
	"crossplane-aws-kinesis-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-kinesis",
	},
	"crossplane-aws-kms": {
		"ghcr.io/crossplane-contrib/provider-aws-kms",
	},
	"crossplane-aws-kms-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-kms",
	},
	"crossplane-aws-lambda": {
		"ghcr.io/crossplane-contrib/provider-aws-lambda",
	},
	"crossplane-aws-lambda-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-lambda",
	},
	"crossplane-aws-rds": {
		"ghcr.io/crossplane-contrib/provider-aws-rds",
	},
	"crossplane-aws-rds-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-rds",
	},
	"crossplane-aws-route53": {
		"ghcr.io/crossplane-contrib/provider-aws-route53",
	},
	"crossplane-aws-route53-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-route53",
	},
	"crossplane-aws-s3": {
		"ghcr.io/crossplane-contrib/provider-aws-s3",
	},
	"crossplane-aws-s3-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-s3",
	},
	"crossplane-aws-sns": {
		"ghcr.io/crossplane-contrib/provider-aws-sns",
	},
	"crossplane-aws-sns-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-sns",
	},
	"crossplane-aws-sqs": {
		"ghcr.io/crossplane-contrib/provider-aws-sqs",
	},
	"crossplane-aws-sqs-fips": {
		"ghcr.io/crossplane-contrib/provider-aws-sqs",
	},
	"cert-manager-acmesolver": {
		"quay.io/jetstack/cert-manager-acmesolver",
	},
	"cert-manager-acmesolver-fips": {
		"quay.io/jetstack/cert-manager-acmesolver",
	},
	"cert-manager-acmesolver-iamguarded": {
		"quay.io/jetstack/cert-manager-acmesolver",
	},
	"cert-manager-acmesolver-iamguarded-fips": {
		"quay.io/jetstack/cert-manager-acmesolver",
	},
	"cert-manager-cainjector": {
		"quay.io/jetstack/cert-manager-cainjector",
	},
	"cert-manager-cainjector-fips": {
		"quay.io/jetstack/cert-manager-cainjector",
	},
	"cert-manager-cainjector-iamguarded": {
		"quay.io/jetstack/cert-manager-cainjector",
	},
	"cert-manager-cainjector-iamguarded-fips": {
		"quay.io/jetstack/cert-manager-cainjector",
	},
	"cert-manager-cmctl": {
		"quay.io/jetstack/cmctl",
	},
	"cert-manager-cmctl-fips": {
		"quay.io/jetstack/cmctl",
	},
	"cert-manager-webhook": {
		"quay.io/jetstack/cert-manager-webhook",
	},
	"cert-manager-webhook-fips": {
		"quay.io/jetstack/cert-manager-webhook",
	},
	"cert-manager-webhook-iamguarded": {
		"quay.io/jetstack/cert-manager-webhook",
	},
	"cert-manager-webhook-iamguarded-fips": {
		"quay.io/jetstack/cert-manager-webhook",
	},
	"flux": {
		"ghcr.io/fluxcd/flux-cli",
	},
	"flux-fips": {
		"ghcr.io/fluxcd/flux-cli",
	},
	"flux-helm-controller": {
		"ghcr.io/fluxcd/helm-controller",
	},
	"flux-helm-controller-fips": {
		"ghcr.io/fluxcd/helm-controller",
	},
	"flux-image-automation-controller": {
		"ghcr.io/fluxcd/image-automation-controller",
	},
	"flux-image-automation-controller-fips": {
		"ghcr.io/fluxcd/image-automation-controller",
	},
	"flux-image-reflector-controller": {
		"ghcr.io/fluxcd/image-reflector-controller",
	},
	"flux-image-reflector-controller-fips": {
		"ghcr.io/fluxcd/image-reflector-controller",
	},
	"flux-kustomize-controller": {
		"ghcr.io/fluxcd/kustomize-controller",
	},
	"flux-kustomize-controller-fips": {
		"ghcr.io/fluxcd/kustomize-controller",
	},
	"flux-notification-controller": {
		"ghcr.io/fluxcd/notification-controller",
	},
	"flux-notification-controller-fips": {
		"ghcr.io/fluxcd/notification-controller",
	},
	"flux-source-controller": {
		"ghcr.io/fluxcd/source-controller",
	},
	"flux-source-controller-fips": {
		"ghcr.io/fluxcd/source-controller",
	},
	"minio-client": {
		"quay.io/minio/mc",
	},
	"minio-client-fips": {
		"quay.io/minio/mc",
	},
	"minio-operator": {
		"quay.io/minio/operator",
	},
	"minio-operator-fips": {
		"quay.io/minio/operator",
	},
	"minio-operator-sidecar": {
		"quay.io/minio/operator-sidecar",
	},
	"minio-operator-sidecar-fips": {
		"quay.io/minio/operator-sidecar",
	},
	"mongodb-kubernetes-operator-readinessprobe": {
		"quay.io/mongodb/mongodb-kubernetes-readinessprobe",
	},
	"mongodb-kubernetes-operator-readinessprobe-fips": {
		"quay.io/mongodb/mongodb-kubernetes-readinessprobe",
	},
	"mongodb-kubernetes-operator-version-upgrade-post-start-hook": {
		"quay.io/mongodb/mongodb-kubernetes-operator-version-upgrade-post-start-hook",
	},
	"mongodb-kubernetes-operator-version-upgrade-post-start-hook-fips": {
		"quay.io/mongodb/mongodb-kubernetes-operator-version-upgrade-post-start-hook",
	},
	"postgres-cloudnative-pg": {
		"ghcr.io/cloudnative-pg/postgresql",
	},
	"postgres-cloudnative-pg-fips": {
		"ghcr.io/cloudnative-pg/postgresql",
	},
	"vault-k8s": {
		"hashicorp/vault-k8s",
	},
}
