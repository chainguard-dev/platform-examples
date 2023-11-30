/*
Copyright 2022 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"flag"
	"log"
	"strings"

	"chainguard.dev/sdk/auth/token"
	common "chainguard.dev/sdk/proto/platform/common/v1"
	iam "chainguard.dev/sdk/proto/platform/iam/v1"
	registry "chainguard.dev/sdk/proto/platform/registry/v1"
	"github.com/google/go-containerregistry/pkg/name"
)

func main() {
	ctx := context.Background()
	tag := flag.String("tag", "cgr.dev/chainguard/static:latest-glibc", "tag to query")
	flag.Parse()

	t, err := name.ParseReference(*tag)
	if err != nil {
		log.Fatalf("parsing tag: %v", err)
	}
	reg := t.Context().RegistryStr()
	fullrepo := t.Context().RepositoryStr()
	group, repo, ok := strings.Cut(fullrepo, "/")
	if !ok {
		log.Fatalf("invalid repo: %s", fullrepo)
	}
	tagstr := t.Identifier()
	log.Printf("registry: %s, group: %s, repo: %s, tag: %s", reg, group, repo, tagstr)

	if reg != "cgr.dev" {
		log.Fatalf("must be in cgr.dev registry")
	}

	// TODO: support cgr.dev/chainguard with anonymous auth.

	// Get the Chainguard auth token.
	audience := "https://console-api.enforce.dev"
	tok, err := token.Load(audience)
	if err != nil {
		log.Fatalf("loading token: %v", err)
	}

	// Set up clients.
	iamc, err := iam.NewClients(ctx, audience, string(tok))
	if err != nil {
		log.Fatalf("creating IAM clients: %v", err)
	}
	regc, err := registry.NewClients(ctx, audience, string(tok))
	if err != nil {
		log.Fatalf("creating Registry clients: %v", err)
	}

	// Get the group UIDP.
	var groupUIDP string
	{
		resp, err := iamc.Groups().List(ctx, &iam.GroupFilter{
			Name: group,
		})
		if err != nil {
			log.Fatalf("listing groups: %v", err)
		}
		if len(resp.Items) != 1 {
			log.Fatalf("expected 1 group, got %d", len(resp.Items))
		}
		groupUIDP = resp.Items[0].Id
	}
	log.Println("group UIDP", groupUIDP)

	// Get the repo UIDP.
	var repoUIDP string
	{
		resp, err := regc.Registry().ListRepos(ctx, &registry.RepoFilter{
			Uidp: &common.UIDPFilter{
				ChildrenOf: groupUIDP,
			},
			Name: repo,
		})
		if err != nil {
			log.Fatalf("listing repos: %v", err)
		}
		if len(resp.Items) != 1 {
			log.Fatalf("expected 1 repo, got %d", len(resp.Items))
		}
		repoUIDP = resp.Items[0].Id
	}
	log.Println("repo UIDP", repoUIDP)

	// Get the tag UIDP.
	var tagUIDP string
	{
		resp, err := regc.Registry().ListTags(ctx, &registry.TagFilter{
			Uidp: &common.UIDPFilter{
				ChildrenOf: repoUIDP,
			},
			Name: tagstr,
		})
		if err != nil {
			log.Fatalf("listing tags: %v", err)
		}
		if len(resp.Items) != 1 {
			log.Fatalf("expected 1 tag, got %d", len(resp.Items))
		}
		tagUIDP = resp.Items[0].Id
	}
	log.Println("tag UIDP", tagUIDP)

	// List tag history for the tag.
	resp, err := regc.Registry().ListTagHistory(ctx, &registry.TagHistoryFilter{
		ParentId: tagUIDP,
	})
	if err != nil {
		log.Fatalf("listing tag history: %v", err)
	}
	for _, i := range resp.Items {
		log.Printf("time: %s digest: %s", i.UpdateTimestamp.AsTime(), i.Digest)
	}
}
