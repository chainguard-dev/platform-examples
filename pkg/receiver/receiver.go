/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package receiver

import (
	"context"
	"log"
	"net/http"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/coreos/go-oidc/v3/oidc"
)

type Handler func(ctx context.Context, event cloudevents.Event) error

// New returns a new Handler that verifies the Event was sent by Chainguard,
// intended for the specified Group, then invokes the provided Handler.
func New(ctx context.Context, issuer, group string, fn Handler) Handler {
	// Construct a verifier that ensures tokens are issued by the Chainguard
	// issuer we expect and are intended for a customer webhook.
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		log.Fatalf("failed to create provider: %v", err)
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: "customer"})

	return func(ctx context.Context, event cloudevents.Event) error {
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
		} else if got := strings.TrimPrefix(tok.Subject, "webhook:"); got != group {
			return cloudevents.NewHTTPResult(http.StatusForbidden, "this token is intended for %s, wanted one for %s", got, group)
		}
		return fn(ctx, event)
	}
}
