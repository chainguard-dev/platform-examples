/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"chainguard.dev/sdk/auth/token"
	common "chainguard.dev/sdk/proto/platform/common/v1"
	iam "chainguard.dev/sdk/proto/platform/iam/v1"
	registry "chainguard.dev/sdk/proto/platform/registry/v1"
	"chainguard.dev/sdk/sts"
	"github.com/google/go-containerregistry/pkg/name"
)

func main() {
	ctx := context.Background()
	var group string
	flag.StringVar(&group, "group", "chainguard", "group to query")
	flag.Parse()

	if len(flag.Args()) != 3 {
		log.Fatalf("requires 3 arguments: repo name, and previous and current image to diff")
	}
	if _, err := name.NewDigest("example.com/foo@" + flag.Arg(1)); err != nil {
		log.Fatalf("invalid digest: %v", err)
	}
	if _, err := name.NewDigest("example.com/foo@" + flag.Arg(2)); err != nil {
		log.Fatalf("invalid digest: %v", err)
	}
	repo, left, right := flag.Arg(0), flag.Arg(1), flag.Arg(2)

	// Get the Chainguard auth token.
	var tok string
	audience := "https://console-api.enforce.dev"
	{
		if group == "chainguard" {
			// This group is special, since anybody can access it by assuming a
			// broadly-assumable identity with permission to view/pull.

			issuer := "https://issuer.enforce.dev"
			resp, err := http.Get("https://justtrustme.dev/token?aud=" + issuer)
			if err != nil {
				log.Fatalf("getting justtrustme token: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Fatalf("getting justtrustme token: %v", resp.Status)
			}
			all, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("reading justtrustme token: %v", err)
			}
			var r struct {
				Token string `json:"token"`
			}
			if err := json.Unmarshal(all, &r); err != nil {
				log.Fatalf("decoding justtrustme token: %v", err)
			}
			tok = r.Token

			tok, err = sts.New(issuer, audience,
				sts.WithIdentity("720909c9f5279097d847ad02a2f24ba8f59de36a/a033a6fabe0bfa0d")).
				Exchange(ctx, tok)
			if err != nil {
				log.Fatalf("exchanging token: %v", err)
			}
		} else {
			if token.RemainingLife(audience, time.Minute) < 0 {
				// TODO: do a browser flow here.
				log.Fatalf("token has expired, please run `chainctl auth login`")
			}
			tokb, err := token.Load(audience)
			if err != nil {
				log.Fatalf("loading token: %v", err)
			}
			tok = string(tokb)
		}
	}

	// Set up clients.
	iamc, err := iam.NewClients(ctx, audience, tok)
	if err != nil {
		log.Fatalf("creating IAM clients: %v", err)
	}
	regc, err := registry.NewClients(ctx, audience, tok)
	if err != nil {
		log.Fatalf("creating Registry clients: %v", err)
	}

	// Get the group UIDP.
	var groupUIDP string
	{
		if group == "chainguard" {
			// This group is special, we'll just hard-code the UIDP.
			groupUIDP = "720909c9f5279097d847ad02a2f24ba8f59de36a"
		} else {
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

	// Get diff for the digests.
	resp, err := regc.Registry().DiffImage(ctx, &registry.DiffImageRequest{
		RepoId:     repoUIDP,
		FromDigest: left,
		ToDigest:   right,
	})
	if err != nil {
		log.Fatalf("diff: %v", err)
	}

	b, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Fatalf("marshaling response: %v", err)
	}
	fmt.Println(string(b))
}
