package main

import (
	"context"
	"flag"
	"log"
	"time"

	"chainguard.dev/sdk/auth/token"
	iam "chainguard.dev/sdk/proto/platform/iam/v1"
	registry "chainguard.dev/sdk/proto/platform/registry/v1"
	"chainguard.dev/sdk/sts"
)

var (
	audience = "https://console-api.enforce.dev"
	issuer   = "https://issuer.enforce.dev"
)

var (
	parent        = flag.String("parent", "", "Organization name")
	repo          = flag.String("repo", "", "Repository name")
	identity      = flag.String("identity", "", "Identity to assume")
	identityToken = flag.String("identity-token", "", "Identity token")
)

func init() {
	flag.Parse()
}

func main() {
	ctx := context.Background()

	// Validate flags
	if *parent == "" {
		log.Fatal("must provide -parent")
	}
	if *repo == "" {
		log.Fatal("must provide -repo")
	}

	// Either assume an identity with an OIDC token, or fetch the token from
	// the location that chainctl saves it to
	var tok string
	if *identity != "" {
		if *identityToken == "" {
			log.Fatal("must provide -identity-token with -identity")
		}
		exch := sts.New("https://issuer.enforce.dev", audience, sts.WithIdentity(*identity))
		cgTok, err := exch.Exchange(ctx, *identityToken)
		if err != nil {
			log.Fatalf("exchanging token: %w", err)
		}

		tok = cgTok.AccessToken
	} else {
		if token.RemainingLife(token.KindAccess, audience, time.Minute) < 0 {
			log.Fatalf("token has expired, please run `chainctl auth login` or provide an identity id and token with -identity and -identity-token")
		}
		tokb, err := token.Load(token.KindAccess, audience)
		if err != nil {
			log.Fatalf("loading token: %v", err)
		}
		tok = string(tokb)
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

	// Get the parent UIDP.
	var parentUIDP string
	resp, err := iamc.Groups().List(ctx, &iam.GroupFilter{
		Name: *parent,
	})
	if err != nil {
		log.Fatalf("listing groups: %v", err)
	}
	if len(resp.Items) != 1 {
		log.Fatalf("expected 1 group, got %d", len(resp.Items))
	}
	parentUIDP = resp.Items[0].Id

	// Create the repository
	createRepoRequest := registry.CreateRepoRequest{
		ParentId: parentUIDP,
		Repo: &registry.Repo{
			Name: *repo,
			SyncConfig: &registry.SyncConfig{
				// Source is the name of the image in the Chainguard
				// catalog that this image is derived from
				Source: *repo,
			},
			// Uncomment this to customize the repository. For
			// instance, by adding a package.
			//
			// See the full spec here:
			//    https://pkg.go.dev/chainguard.dev/sdk/proto/platform/registry/v1#CustomOverlay
			//CustomOverlay: &registry.CustomOverlay{
			//	Contents: &registry.ImageContents{
			//		Packages: []string{
			//			"curl",
			//		},
			//	},
			//},
		},
		// Returns an error, rather than updating the repo, if it
		// already exists.
		PreventExisting: true,
	}
	createdRepo, err := regc.Registry().CreateRepo(ctx, &createRepoRequest)
	if err != nil {
		log.Fatalf("creating repository: %v", err)
	}

	log.Printf("Created repository; name=%s id=%s", createdRepo.Name, createdRepo.Id)
}
