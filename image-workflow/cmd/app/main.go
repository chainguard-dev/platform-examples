/*
Copyright 2022 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/go-github/v43/github"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
)

type envConfig struct {
	Issuer           string `envconfig:"ISSUER_URL" required:"true"`
	Group            string `envconfig:"GROUP" required:"true"`
	Port             int    `envconfig:"PORT" default:"8080" required:"true"`
	GithubOrg        string `envconfig:"GITHUB_ORG" required:"true"`
	GithubRepo       string `envconfig:"GITHUB_REPO" required:"true"`
	GithubToken      string `envconfig:"GITHUB_TOKEN" required:"true"`
	GithubWorkflowID string `envconfig:"GITHUB_WORKFLOW_ID" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("failed to process env var: %s", err)
	}

	client := github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.TrimSpace(env.GithubToken)},
	)))

	c, err := cloudevents.NewClientHTTP(cloudevents.WithPort(env.Port),
		// We need to infuse the request onto context, so we can
		// authenticate requests.
		cehttp.WithRequestDataAtContextMiddleware())
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

	receiver := func(ctx context.Context, event cloudevents.Event) error {
		// We expect Chainguard webhooks to pass an Authorization header.
		auth := strings.TrimPrefix(cehttp.RequestDataFromContext(ctx).Header.Get("Authorization"), "Bearer ")
		if auth == "" {
			return cloudevents.NewHTTPResult(http.StatusUnauthorized, "Unauthorized")
		}

		// Verify that the token is well-formed, and in fact intended for us!
		if tok, err := verifier.Verify(ctx, auth); err != nil {
			return cloudevents.NewHTTPResult(http.StatusForbidden, "unable to verify token: %w", err)
		} else if !strings.HasPrefix(tok.Subject, "webhook:") {
			return cloudevents.NewHTTPResult(http.StatusForbidden, "subject should be from the Chainguard webhook component, got: %s", tok.Subject)
		} else if group := strings.TrimPrefix(tok.Subject, "webhook:"); group != env.Group {
			return cloudevents.NewHTTPResult(http.StatusForbidden, "this token is intended for %s, wanted one for %s", group, env.Group)
		}

		// We are handling a specific event type, so filter the rest.
		if event.Type() != PushEventType {
			return nil
		}

		data := Occurrence{}
		if err := event.DataAs(&data); err != nil {
			return cloudevents.NewHTTPResult(http.StatusInternalServerError, "unable to unmarshal data: %w", err)
		}

		log.Printf("got event: %+v", data)

		img := "cgr.dev/" + data.Body.Repository

		log.Println("got event for image:", img)
		if _, err := client.Actions.CreateWorkflowDispatchEventByFileName(ctx, env.GithubOrg, env.GithubRepo, env.GithubWorkflowID, github.CreateWorkflowDispatchEventRequest{
			Ref: "main",
			Inputs: map[string]interface{}{
				"image": img,
			},
		}); err != nil {
			log.Printf("failed to dispatch workflow: %v", err)
			return cloudevents.NewHTTPResult(http.StatusInternalServerError, "failed to dispatch workflow: %w", err)
		}
		log.Printf("dispatched workflow %s (image=%q)", env.GithubWorkflowID, img)

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
