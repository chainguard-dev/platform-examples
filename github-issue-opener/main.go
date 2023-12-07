/*
Copyright 2022 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"chainguard.dev/sdk/events"
	"chainguard.dev/sdk/events/policy"
	"chainguard.dev/sdk/events/receiver"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/google/go-github/v43/github"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
)

type envConfig struct {
	Issuer      string   `envconfig:"ISSUER_URL" required:"true"`
	Group       string   `envconfig:"GROUP" required:"true"`
	Port        int      `envconfig:"PORT" default:"8080" required:"true"`
	GithubOrg   string   `envconfig:"GITHUB_ORG" required:"true"`
	GithubRepo  string   `envconfig:"GITHUB_REPO" required:"true"`
	GithubToken string   `envconfig:"GITHUB_TOKEN" required:"true"`
	Labels      []string `envconfig:"LABELS" required:"false"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("failed to process env var: %s", err)
	}
	ctx := context.Background()

	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.TrimSpace(env.GithubToken)},
	)))

	receiver, err := receiver.New(ctx, env.Issuer, env.Group, func(ctx context.Context, event cloudevents.Event) error {
		// We are handling a specific event type, so filter the rest.
		if event.Type() != policy.ChangedEventType {
			return nil
		}

		body := &policy.ImagePolicyRecord{}
		data := events.Occurrence{Body: body}
		if err := event.DataAs(&data); err != nil {
			return cloudevents.NewHTTPResult(http.StatusBadRequest, "unable to unmarshal data: %w", err)
		}
		for name, pol := range body.Policies {
			if pol.Valid {
				// Not in violation of policy
				continue
			}
			switch pol.Change {
			case policy.ImprovedChange:
				// TODO: How is this an improvement?
				continue
			case policy.NewChange, policy.DegradedChange:
				// We want to fire on these events.
			}

			issue, _, err := client.Issues.Create(ctx, env.GithubOrg, env.GithubRepo, &github.IssueRequest{
				Title:  ptr(fmt.Sprintf("Policy %s failed", name)),
				Labels: &env.Labels,
				Body: ptr(strings.Join([]string{
					fmt.Sprintf("Image:        `%s`", body.ImageID),
					fmt.Sprintf("Cluster       `%s`", body.ClusterID),
					fmt.Sprintf("Policy:       `%s`", name),
					fmt.Sprintf("Last Checked: `%v`", pol.LastChecked.Time),
					fmt.Sprintf("Diagnostic:   `%v`", pol.Diagnostic),
				}, "\n")),
			})
			if err != nil {
				return cloudevents.NewHTTPResult(http.StatusInternalServerError, "unable to create GitHub issue: %w", err)
			}
			log.Printf("Opened issue: %d", issue.GetNumber())
		}

		return nil
	})
	if err != nil {
		log.Fatalf("failed to create receiver: %v", err)
	}

	c, err := cloudevents.NewClientHTTP(cloudevents.WithPort(env.Port),
		// We need to infuse the request onto context, so we can
		// authenticate requests.
		cehttp.WithRequestDataAtContextMiddleware())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
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

func ptr(s string) *string {
	return &s
}
