package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"chainguard.dev/sdk/events/receiver"
	"chainguard.dev/sdk/events/registry"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/api/idtoken"
)

func main() {
	var env struct {
		Issuer     string `envconfig:"ISSUER_URL" required:"true"`
		Group      string `envconfig:"GROUP" required:"true"`
		Port       int    `envconfig:"PORT" default:"8080" required:"true"`
		IngressURI string `envconfig:"EVENT_INGRESS_URI" required:"true"`
	}
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	c, err := idtoken.NewClient(ctx, env.IngressURI)
	if err != nil {
		log.Fatalf("failed to create idtoken client: %v", err) //nolint:gocritic
	}
	ceclient, err := cloudevents.NewClientHTTP(
		cloudevents.WithTarget(env.IngressURI),
		cehttp.WithClient(http.Client{Transport: c.Transport}))
	if err != nil {
		log.Fatalf("failed to create cloudevents client: %v", err)
	}

	receiver, err := receiver.New(ctx, env.Issuer, env.Group, func(ctx context.Context, event cloudevents.Event) error {
		// We are handling a specific event type, so filter the rest.
		if event.Type() != registry.PulledEventType {
			return nil
		}

		// Forward the event.
		const retryDelay = 10 * time.Millisecond
		const maxRetry = 3
		rctx := cloudevents.ContextWithRetriesExponentialBackoff(context.WithoutCancel(ctx), retryDelay, maxRetry)
		if ceresult := ceclient.Send(rctx, event); cloudevents.IsUndelivered(ceresult) || cloudevents.IsNACK(ceresult) {
			return fmt.Errorf("failed to forward event: %w", ceresult)
		}
		log.Println("event forwarded")
		return nil
	})
	if err != nil {
		log.Fatalf("failed to create receiver: %v", err)
	}

	ceserver, err := cloudevents.NewClientHTTP(cloudevents.WithPort(env.Port),
		// We need to infuse the request onto context, so we can
		// authenticate requests.
		cehttp.WithRequestDataAtContextMiddleware())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	if err := ceserver.StartReceiver(ctx, func(ctx context.Context, event cloudevents.Event) error {
		// This thunk simply wraps the main receiver in one that logs any errors
		// we encounter.
		if err := receiver(ctx, event); err != nil {
			log.Printf("Error handling event: %v", err)
			return err
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}
