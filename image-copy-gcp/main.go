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

	"chainguard.dev/sdk/events"
	"chainguard.dev/sdk/events/receiver"
	"chainguard.dev/sdk/events/registry"
	v1 "chainguard.dev/sdk/proto/platform/registry/v1"
	"chainguard.dev/sdk/sts"
	"cloud.google.com/go/compute/metadata"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/api/idtoken"
)

type envConfig struct {
	APIEndpoint string `envconfig:"API_ENDPOINT" required:"true"`
	Issuer      string `envconfig:"ISSUER_URL" required:"true"`
	GroupName   string `envconfig:"GROUP_NAME" required:"true"`
	Group       string `envconfig:"GROUP" required:"true"`
	Identity    string `envconfig:"IDENTITY" required:"true"`
	Port        int    `envconfig:"PORT" default:"8080" required:"true"`
	DstRepo     string `envconfig:"DST_REPO" required:"true"` // Almost fully qualified at this point, just needs the final component.
}

var location, sa string
var srcRepo name.Repository
var env envConfig

func init() {
	var err error
	location, err = metadata.Zone()
	if err != nil {
		log.Panicf("getting location: %v", err)
	}
	sa, err = metadata.Email("default")
	if err != nil {
		log.Panicf("getting SA: %v", err)
	}
	if err := envconfig.Process("", &env); err != nil {
		log.Panicf("failed to process env var: %s", err)
	}
}

func main() {
	log.Printf("env: %+v", env)
	log.Printf("location: %s", location)
	log.Printf("sa: %s", sa)

	ctx := context.Background()
	receiver, err := receiver.New(ctx, env.Issuer, env.Group, func(ctx context.Context, event cloudevents.Event) error {
		// We are handling a specific event type, so filter the rest.
		if event.Type() != registry.PushedEventType {
			return nil
		}

		body := registry.PushEvent{}
		data := events.Occurrence{Body: &body}
		if err := event.DataAs(&data); err != nil {
			return cloudevents.NewHTTPResult(http.StatusBadRequest, "unable to unmarshal data: %w", err)
		}
		log.Printf("got event: %+v", data)

		// Check that the event is one we care about:
		// - It's not a push error.
		// - It's a tag push.
		if body.Error != nil {
			log.Printf("event body has error, skipping: %+v", body.Error)
			return nil
		}
		if body.Tag == "" || body.Type != "manifest" {
			log.Printf("event body is not a tag push, skipping: %q %q", body.Tag, body.Type)
			return nil
		}

		// Resolve the repository ID to the name
		repoName, err := resolveRepositoryName(ctx, body.RepoID)
		if err != nil {
			return fmt.Errorf("failed to resolve repository name from id in the event: %w", err)
		}

		src := "cgr.dev/" + env.GroupName + "/" + repoName + ":" + body.Tag
		dst := env.DstRepo + "/" + repoName + ":" + body.Tag
		log.Printf("Copying %s to %s...", src, dst)
		if err := crane.Copy(src, dst,
			crane.WithAuthFromKeychain(authn.NewMultiKeychain(
				google.Keychain,
				cgKeychain{},
			))); err != nil {
			return fmt.Errorf("copying image: %w", err)
		}
		log.Println("Copied!")
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

func resolveRepositoryName(ctx context.Context, repoID string) (string, error) {
	// Generate a token for the Chainguard API
	tok, err := newToken(ctx, env.APIEndpoint)
	if err != nil {
		return "", fmt.Errorf("getting token: %w", err)
	}

	// Create client that uses the token
	client, err := v1.NewClients(ctx, env.APIEndpoint, tok.AccessToken)
	if err != nil {
		return "", fmt.Errorf("creating clients: %w", err)
	}

	// Lookup the repository name from the ID
	repoList, err := client.Registry().ListRepos(ctx, &v1.RepoFilter{
		Id: repoID,
	})
	if err != nil {
		return "", fmt.Errorf("listing repositories: %w", err)
	}
	for _, repo := range repoList.Items {
		return repo.Name, nil
	}

	return "", fmt.Errorf("couldn't find repository name for id: %s", repoID)
}

func newToken(ctx context.Context, audience string) (*sts.TokenPair, error) {
	exch := sts.New(env.Issuer, audience, sts.WithIdentity(env.Identity))
	ts, err := idtoken.NewTokenSource(ctx, env.Issuer)
	if err != nil {
		return nil, fmt.Errorf("getting token source: %w", err)
	}
	tok, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("getting token: %w", err)
	}
	cgTok, err := exch.Exchange(ctx, tok.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("exchanging token: %w", err)
	}

	return &cgTok, nil
}

type cgKeychain struct{}

func (k cgKeychain) Resolve(res authn.Resource) (authn.Authenticator, error) {
	if res.RegistryStr() != "cgr.dev" {
		return authn.Anonymous, nil
	}

	tok, err := newToken(context.Background(), res.RegistryStr())
	if err != nil {
		return nil, fmt.Errorf("getting token: %w", err)
	}

	return &authn.Basic{
		Username: "_token",
		Password: tok.AccessToken,
	}, nil
}
