/*
Copyright 2022 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/go-github/v43/github"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
	"knative.dev/pkg/ptr"

	"chainguard.dev/api/pkg/events"
	"chainguard.dev/api/proto/platform/iam"
)

type envConfig struct {
	Issuer      string `envconfig:"ISSUER_URL" required:"true"`
	Group       string `envconfig:"GROUP" required:"true"`
	Port        int    `envconfig:"PORT" default:"8080" required:"true"`
	GithubOrg   string `envconfig:"GITHUB_ORG" required:"true"`
	GithubRepo  string `envconfig:"GITHUB_Repo" required:"true"`
	GithubToken string `envconfig:"GITHUB_TOKEN" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("failed to process env var: %s", err)
	}
	c, err := cloudevents.NewClientHTTP(cloudevents.WithPort(env.Port),
		// We need to infuse the request onto context, so we can
		// authenticate requests.
		http.WithRequestDataAtContextMiddleware())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	ctx := context.Background()

	// Construct a verifier that ensures tokens are issued by the Chainguard
	// issuer we expect and are intended for a customer webhook.
	provider, err := oidc.NewProvider(ctx, env.Issuer)
	if err != nil {
		log.Fatalf("failed to create provider: %v", err)
	}
	verifier := provider.Verifier(&oidc.Config{
		ClientID: "customer",
	})

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: env.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	receiver := func(ctx context.Context, event cloudevents.Event) error {
		// We expect Chainguard webhooks to pass an Authorization header.
		auth := strings.TrimPrefix(http.RequestDataFromContext(ctx).Header.Get("Authorization"), "Bearer ")
		if auth == "" {
			// TODO: Return a 401
			return errors.New("Unauthorized")
		}

		// Verify that the token is well-formed, and in fact intended for us!
		if tok, err := verifier.Verify(ctx, auth); err != nil {
			// TODO: Return a 403
			return fmt.Errorf("unable to verify token: %w", err)
		} else if !strings.HasPrefix(tok.Subject, "webhook:") {
			// TODO: Return a 403
			return fmt.Errorf("subject should be from the Chainguard webhook component, got: %s", tok.Subject)
		} else if group := strings.TrimPrefix(tok.Subject, "webhook:"); group != env.Group {
			// TODO: Return a 403
			return fmt.Errorf("this token is intended for %s, wanted one for %s", group, env.Group)
		}

		// We are handling a specific event type, so filter the rest.
		// TODO: Replace this with whatever type we use for Continuous
		// Verification policy violations.
		if event.Type() != "dev.chainguard.api.iam.group_invite.created.v1" {
			return nil
		}

		// TODO: Replace this with whatever type we use for Continuous
		// Verification policy violations.
		body := &iam.GroupInvite{}
		data := events.Occurrence{
			Body: body,
		}
		if err := event.DataAs(&data); err != nil {
			// TODO: Return a 500
			return fmt.Errorf("unable to unmarshal data: %w", err)
		}

		issue, _, err := client.Issues.Create(ctx, env.GithubOrg, env.GithubRepo, &github.IssueRequest{
			// TODO: Replace this with an issue title/body based on the
			// Continuous Verification payload, once we have it.
			Title: ptr.String("This is the title"),
			Body:  ptr.String(fmt.Sprintf("This is the body %s", body.Id)),
		})
		if err != nil {
			return err
		}

		log.Printf("Opened issue: %d", issue.GetNumber())
		return nil
	}

	if err := c.StartReceiver(ctx, func(ctx context.Context, event cloudevents.Event) error {
		// This thunk simply wraps the main receiver in one that logs any errors
		// we encounter.
		err := receiver(ctx, event)
		if err != nil {
			log.Printf("SAW: %v", err)
		}
		return err
	}); err != nil {
		log.Fatal(err)
	}
}
