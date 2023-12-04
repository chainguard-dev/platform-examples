/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"

	"chainguard.dev/sdk/auth/aws"
	common "chainguard.dev/sdk/proto/platform/common/v1"
	registry "chainguard.dev/sdk/proto/platform/registry/v1"
	"chainguard.dev/sdk/sts"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/kelseyhightower/envconfig"
)

var env = struct {
	APIEndpoint string `envconfig:"API_ENDPOINT" required:"true"`
	Issuer      string `envconfig:"ISSUER_URL" required:"true"`
	Group       string `envconfig:"GROUP" required:"true"`
	Identity    string `envconfig:"IDENTITY" required:"true"`
}{}

func init() {
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("failed to process env var: %s", err)
	}
}
func main() { lambda.Start(handler) }

func handler(ctx context.Context, levent events.LambdaFunctionURLRequest) (resp string, err error) {
	// Get AWS credentials.
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load configuration, %w", err)
	}
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve credentials, %w", err)
	}

	// Generate a token and exchange it for a Chainguard token.
	awsTok, err := aws.GenerateToken(ctx, creds, env.Issuer, env.Identity)
	if err != nil {
		return "", fmt.Errorf("generating AWS token: %w", err)
	}
	exch := sts.New(env.Issuer, env.APIEndpoint, sts.WithIdentity(env.Identity))
	cgtok, err := exch.Exchange(ctx, awsTok)
	if err != nil {
		return "", fmt.Errorf("exchanging token: %w", err)
	}

	// Use the token to list repos in the group.
	clients, err := registry.NewClients(ctx, env.APIEndpoint, cgtok)
	if err != nil {
		return "", fmt.Errorf("creating clients: %w", err)
	}
	ls, err := clients.Registry().ListRepos(ctx, &registry.RepoFilter{
		Uidp: &common.UIDPFilter{
			ChildrenOf: env.Group,
		},
	})
	if err != nil {
		return "", fmt.Errorf("listing repos: %w", err)
	}
	return fmt.Sprintf("repos: %v", ls.Items), nil
}
