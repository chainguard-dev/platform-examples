/*
Copyright 2022 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"cloud.google.com/go/compute/metadata"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/api/idtoken"
)

type envConfig struct {
	Issuer   string `envconfig:"ISSUER_URL" required:"true"`
	Group    string `envconfig:"GROUP" required:"true"`
	Identity string `envconfig:"IDENTITY" required:"true"`
	Port     int    `envconfig:"PORT" default:"8080" required:"true"`
	DstRepo  string `envconfig:"DST_REPO" required:"true"` // Almost fully qualified at this point, just needs the final component.
}

var location, sa string
var srcRepo name.Repository

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
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("failed to process env var: %s", err)
	}
	log.Printf("env: %+v", env)
	log.Printf("location: %s", location)
	log.Printf("sa: %s", sa)

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

		src := "cgr.dev/" + data.Body.Repository
		dst := env.DstRepo + "/" + filepath.Base(data.Body.Repository)
		log.Printf("Copying %s to %s...", src, dst)
		if err := crane.Copy(src, dst,
			crane.WithAuthFromKeychain(authn.NewMultiKeychain(
				google.Keychain,
				cgKeychain{env.Issuer, env.Identity},
			))); err != nil {
			return fmt.Errorf("copying image: %w", err)
		}
		log.Println("Copied!")
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

type cgKeychain struct {
	issuer, identity string
}

func (k cgKeychain) Resolve(res authn.Resource) (authn.Authenticator, error) {
	if res.RegistryStr() != "cgr.dev" {
		return authn.Anonymous, nil
	}

	url := fmt.Sprintf("%s/sts/exchange?aud=%s&identity=%s", k.issuer, res.RegistryStr(), k.identity)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	client, err := idtoken.NewClient(context.Background(), k.issuer)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	all, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got HTTP %d to /sts/exchange: %s", resp.StatusCode, all)
	}
	var m map[string]string
	if err := json.NewDecoder(bytes.NewReader(all)).Decode(&m); err != nil {
		return nil, err
	}
	return &authn.Basic{
		Username: "_token",
		Password: m["token"],
	}, nil
}
