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

	admissionv1 "k8s.io/api/admission/v1"

	"chainguard.dev/sdk/pkg/events"
	"chainguard.dev/sdk/pkg/events/admission"
	"chainguard.dev/sdk/pkg/events/policy"
	"chainguard.dev/sdk/pkg/events/receiver"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kelseyhightower/envconfig"
	"github.com/slack-go/slack"
)

const (
	warn  string = "WARN"
	info  string = "INFO"
	debug string = "DEBUG"
)

type envConfig struct {
	Console      string `envconfig:"CONSOLE_URL" required:"true"`
	Issuer       string `envconfig:"ISSUER_URL" required:"true"`
	Group        string `envconfig:"GROUP" required:"true"`
	Port         int    `envconfig:"PORT" default:"8080" required:"true"`
	SlackWebhook string `envconfig:"SLACK_WEBHOOK" required:"true"`
	NotifyLevel  string `envconfig:"NOTIFY_LEVEL" required:"true"`
	Debug        bool   `envconfig:"DEBUG" default:"false" required:"false"`
}

func main() {
	log.Printf("Starting Slack Webhook")
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("failed to process env var: %s", err)
	}

	// TODO(mattmoor): Validate that env.Group is a valid UIDP once our library
	// is public, so that we can fail faster on user error.

	log.Printf("Console URL: %v", env.Console)
	log.Printf("Group veiwing events: %v", env.Group)
	log.Printf("Sending events %v", env.SlackWebhook)
	log.Printf("Issuer: %v", env.Issuer)
	log.Printf("Notify Level: %v", env.NotifyLevel)

	slackHttpClient := &http.Client{Transport: newAddHeaderTransport(nil)}

	ctx := context.Background()

	receiver, err := receiver.New(ctx, env.Issuer, env.Group, func(ctx context.Context, event cloudevents.Event) error {
		log.Printf("Processing Event Type: %v", event.Type())

		switch EventType := event.Type(); EventType {
		case policy.ChangedEventType:
			var ipr = policy.ImagePolicyRecord{}
			occ := events.Occurrence{
				Body: &ipr,
			}
			if err := event.DataAs(&occ); err != nil {
				return cloudevents.NewHTTPResult(http.StatusInternalServerError, "unable to unmarshal data: %w", err)
			}
			log.Printf("Image Policy Cluster ID: %v", ipr.ClusterID)

			msg := env.imagePolicyRecordToWebhookMessage(ipr)
			if err := slack.PostWebhookCustomHTTP(env.SlackWebhook, slackHttpClient, msg); err != nil {
				return cloudevents.NewHTTPResult(http.StatusInternalServerError, "unable to send to slack webhook: %w", err)
			}
			return nil

		case admission.ReviewEventType:
			admission := admissionv1.AdmissionReview{}
			occ := events.Occurrence{
				Body: &admission,
			}
			if err := event.DataAs(&occ); err != nil {
				return cloudevents.NewHTTPResult(http.StatusInternalServerError, "unable to unmarshal data: %w", err)
			}
			log.Printf("Response Message %v", admission.Response.Result.Message)

			msg := env.admissionReviewToWebhookMessage(admission)
			if err := slack.PostWebhookCustomHTTP(env.SlackWebhook, slackHttpClient, msg); err != nil {
				return cloudevents.NewHTTPResult(http.StatusInternalServerError, "unable to send to slack webhook: %w", err)
			}
			return nil
		default:
			if env.Debug {
				log.Printf("EventType:%v", EventType)
				log.Printf("Event Body: %v", string(event.Data()))
			}
			return nil
		}
	})
	if err != nil {
		log.Fatalf("failed to create receiver: %v", err)
	}

	c, err := cloudevents.NewClientHTTP(cloudevents.WithPort(env.Port),
		cloudevents.WithHeader("User-Agent", "Chainguard Enforce"),
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

func (e *envConfig) imagePolicyRecordToWebhookMessage(ipr policy.ImagePolicyRecord) *slack.WebhookMessage {
	divSection := slack.NewDividerBlock()

	// Header Section
	headerText := slack.NewTextBlockObject("mrkdwn",
		fmt.Sprintf("*Policy Alert* from _<%s/clusters/%s|%s>_ related to image %s:", e.Console, ipr.ClusterID, ipr.ClusterID, ipr.ImageID),
		false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	blocks := &slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
		},
	}

	out := 0
	for name, state := range ipr.Policies {
		var valid string
		var emoji string
		if state.Valid {
			valid = "passing"
		} else {
			valid = "failing"
		}

		var body string
		switch state.Change {
		case policy.NewChange:
			if state.Valid {
				emoji = ":star:"
			} else {
				emoji = ":x:"
			}
			body = fmt.Sprintf("\t%s [%s] Policy _%s_ now applies and is *%s*", emoji, name, name, valid)
		case policy.DegradedChange:
			emoji = ":fire:"
			body = fmt.Sprintf("\t%s [%s] Degraded change detected for policy _%s_ and is now *%s*.", emoji, name, name, valid)
		case policy.ImprovedChange:
			emoji = ":star-struck:"
			body = fmt.Sprintf("\t%s [%s] Improved change detected for policy _%s_ and is now *%s*.", emoji, name, name, valid)
		default:
			// No change, don't report.
			continue
		}
		if state.Diagnostic != "" {
			body += "\n\n```\n" + state.Diagnostic + "\n```\n"
		}
		stateText := slack.NewTextBlockObject("mrkdwn", body, false, false)

		if e.shouldFilterNotification(state) {
			log.Printf("Not notifying %q due to notify level: %s", stateText.Text, e.NotifyLevel)
			continue
		}

		blocks.BlockSet = append(blocks.BlockSet, slack.NewSectionBlock(stateText, nil, nil))
		out++
	}

	// If we did not add any blocks, don't return the webhook message.
	if out == 0 {
		return nil
	}
	blocks.BlockSet = append(blocks.BlockSet, divSection)
	return &slack.WebhookMessage{
		Blocks: blocks,
	}
}

func (e *envConfig) admissionReviewToWebhookMessage(adm admissionv1.AdmissionReview) *slack.WebhookMessage {
	divSection := slack.NewDividerBlock()

	user := adm.Request.UserInfo.Username
	podName := adm.Request.Name
	namespace := adm.Request.Namespace
	message := adm.Response.Result.Message

	// Header Section
	headerText := slack.NewTextBlockObject("mrkdwn",
		"*Admission Alert*", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	blocks := &slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
		},
	}

	emoji := ":fire:"

	body := fmt.Sprintf("\t%s User %v tried to deploy Pod %s in Namespace %v but failed because of:\n```\n%s\n```", emoji, user, podName, namespace, message)

	stateText := slack.NewTextBlockObject("mrkdwn", body, false, false)
	blocks.BlockSet = append(blocks.BlockSet, slack.NewSectionBlock(stateText, nil, nil))
	blocks.BlockSet = append(blocks.BlockSet, divSection)
	return &slack.WebhookMessage{
		Blocks: blocks,
	}

}

func (e *envConfig) shouldFilterNotification(state *policy.State) bool {
	switch e.NotifyLevel {
	case warn:
		// Filter out improvement changes
		if policy.ImprovedChange == state.Change {
			return true
		}
		fallthrough
	case info:
		// Filter out new passing
		if state.Valid && policy.NewChange == state.Change {
			return true
		}
		fallthrough
	case debug:
		// Always log
		fallthrough
	default:
		return false
	}
}

type addHeaderTransport struct {
	T http.RoundTripper
}

func (adt *addHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "Enforce-Events")
	return adt.T.RoundTrip(req)
}

func newAddHeaderTransport(T http.RoundTripper) *addHeaderTransport {
	if T == nil {
		T = http.DefaultTransport
	}
	return &addHeaderTransport{T}
}
