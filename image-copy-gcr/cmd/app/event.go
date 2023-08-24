/*
Copyright 2022 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

// NOTE: these types will eventually be made available as part of a Chainguard
// SDK along with our API clients.

// Occurrence is the CloudEvent payload for events.
type Occurrence struct {
	Actor *Actor `json:"actor,omitempty"`

	// Body is the resource that was created.
	// For this sample, it will always be RegistryPush.
	Body RegistryPush `json:"body,omitempty"`
}

// Actor is the event payload form of which identity was responsible for the
// event.
type Actor struct {
	// Subject is the identity that triggered this event.
	Subject string `json:"subject"`

	// Actor contains the name/value pairs for each of the claims that were
	// validated to assume the identity whose UIDP appears in Subject above.
	Actor map[string]string `json:"act,omitempty"`
}

// ChangedEventType is the cloudevents event type for registry push events.
const PushEventType = "dev.chainguard.registry.push.v1"

// RegistryPush describes an item being pushed to the registry.
type RegistryPush struct {
	// Repository identifies the repository being pushed
	Repository string `json:"repository"`

	// Tag holds the tag being pushed, if there is one.
	Tag string `json:"tag,omitempty"`

	// Digest holds the digest being pushed.
	// Digest will hold the sha256 of the content being pushed, whether that is
	// a blob or a manifest.
	Digest string `json:"digest"`

	// Type determines whether the object being pushed is a manifest or blob.
	Type string `json:"type"`

	// When holds when the push occurred.
	//When civil.DateTime `json:"when"`

	// Location holds the detected approximate location of the client who pulled.
	// For example, "ColumbusOHUS" or "Minato City13JP".
	Location string `json:"location"`

	// UserAgent holds the user-agent of the client who pulled.
	UserAgent string `json:"user_agent" bigquery:"user_agent"`

	Error *Error `json:"error,omitempty"`
}

type Error struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
