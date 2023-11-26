/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"chainguard.dev/sdk/pkg/auth/aws"
	cgevents "chainguard.dev/sdk/pkg/events"
	"chainguard.dev/sdk/pkg/events/registry"
	"chainguard.dev/sdk/pkg/sts"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	ecrcreds "github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/coreos/go-oidc"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/kelseyhightower/envconfig"
)

var amazonKeychain authn.Keychain = authn.NewKeychainFromHelper(ecrcreds.NewECRHelper(ecrcreds.WithLogger(log.Writer())))

var env = struct {
	Issuer        string `envconfig:"ISSUER_URL" required:"true"`
	Group         string `envconfig:"GROUP" required:"true"`
	Identity      string `envconfig:"IDENTITY" required:"true"`
	Region        string `envconfig:"REGION" required:"true"`
	DstRepo       string `envconfig:"DST_REPO" required:"true"`
	FullDstRepo   string `envconfig:"FULL_DST_REPO" required:"true"`
	ImmutableTags bool   `envconfig:"IMMUTABLE_TAGS" required:"true"`
}{}

func init() {
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("failed to process env var: %s", err)
	}
}
func main() { lambda.Start(handler) }

func handler(ctx context.Context, levent events.LambdaFunctionURLRequest) (resp string, err error) {
	defer func() {
		if err != nil {
			log.Printf("=== GOT ERROR: %v", err)
			log.Printf("body: %+v", levent.Body)
			log.Printf("env: %+v", env)
		}
	}()

	// Construct a verifier that ensures tokens are issued by the Chainguard
	// issuer we expect and are intended for us.
	auth := strings.TrimPrefix(levent.Headers["authorization"], "Bearer ")
	if auth == "" {
		return "", fmt.Errorf("auth header missing")
	}
	provider, err := oidc.NewProvider(ctx, env.Issuer)
	if err != nil {
		return "", fmt.Errorf("failed to create provider: %v", err)
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: "customer"})
	if tok, err := verifier.Verify(ctx, auth); err != nil {
		return "", fmt.Errorf("unable to verify token: %w", err)
	} else if !strings.HasPrefix(tok.Subject, "webhook:") {
		return "", fmt.Errorf("subject should be from the Chainguard webhook component, got: %s", tok.Subject)
	} else if group := strings.TrimPrefix(tok.Subject, "webhook:"); group != env.Group {
		return "", fmt.Errorf("this token is intended for %s, wanted one for %s", group, env.Group)
	}

	// Check that the event is one we care about:
	// - It's a registry push event.
	// - It's not a push error.
	// - It's a tag push.
	if levent.Headers["ce-type"] != registry.PushedEventType {
		log.Printf("event type is %q, skipping", levent.Headers["ce-type"])
		return "", nil
	}
	body := registry.PushEvent{}
	data := cgevents.Occurrence{Body: &body}
	if err := json.Unmarshal([]byte(levent.Body), &data); err != nil {
		return "", fmt.Errorf("unable to unmarshal event: %w", err)
	}
	if body.Error != nil {
		log.Printf("event body has error, skipping: %+v", body.Error)
		return "", nil
	}
	if body.Tag == "" || body.Type != "manifest" {
		log.Printf("event body is not a tag push, skipping: %q %q", body.Tag, body.Type)
		return "", nil
	}

	// Attempt to create the repo; if it exists, ignore it.
	// ECR requires you to pre-create repos before pushing to them.
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load configuration, %w", err)
	}
	repo := filepath.Join(env.DstRepo, filepath.Base(body.Repository))
	tagMutability := types.ImageTagMutabilityMutable
	if env.ImmutableTags {
		tagMutability = types.ImageTagMutabilityImmutable
	}
	if _, err := ecr.New(ecr.Options{
		Region:      env.Region,
		Credentials: cfg.Credentials,
	}).CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName:     &repo,
		ImageTagMutability: tagMutability,
	}); err != nil {
		var rae *types.RepositoryAlreadyExistsException
		if errors.As(err, &rae) {
			log.Printf("ECR repo %s already exists", repo)
		} else {
			return "", fmt.Errorf("creating ECR repo %s: %w", repo, err)
		}
	} else {
		log.Printf("Created ECR repo %s", repo)
	}

	// Sync src:tag to dst:tag.
	src := "cgr.dev/" + body.Repository + ":" + body.Tag
	dst := filepath.Join(env.FullDstRepo, filepath.Base(body.Repository)) + ":" + body.Tag
	kc := authn.NewMultiKeychain(
		amazonKeychain,
		cgKeychain{env.Issuer, env.Region, env.Identity},
	)
	if env.ImmutableTags {
		dig, err := crane.Digest(src, crane.WithAuthFromKeychain(kc))
		if err != nil {
			return "", fmt.Errorf("getting digest for %s: %w", src, err)
		}
		dst += "-" + strings.TrimPrefix(dig, "sha256:")[:6]
	}
	log.Printf("Copying %s to %s...", src, dst)
	if err := crane.Copy(src, dst, crane.WithAuthFromKeychain(kc)); err != nil {
		return "", fmt.Errorf("copying image: %w", err)
	}
	log.Printf("Copied %s to %s", src, dst)
	return "", nil
}

// cgKeychain is an authn.Keychain that provides a Chainguard token capable of pulling from cgr.dev.
//
// It does this by first generating an AWS token, then exchanging it for a Chainguard token for the
// specified Chainguard identity, which has been set up to be assumed by the AWS identity.
type cgKeychain struct {
	issuer, region, identity string
}

func (k cgKeychain) Resolve(res authn.Resource) (authn.Authenticator, error) {
	if res.RegistryStr() != "cgr.dev" {
		return authn.Anonymous, nil
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration, %w", err)
	}
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials, %w", err)
	}

	awsTok, err := aws.GenerateToken(ctx, creds, k.issuer, k.identity)
	if err != nil {
		return nil, fmt.Errorf("generating AWS token: %w", err)
	}

	exch := sts.New(k.issuer, res.RegistryStr(), sts.WithIdentity(k.identity))
	cgtok, err := exch.Exchange(ctx, awsTok)
	if err != nil {
		return nil, fmt.Errorf("exchanging token: %w", err)
	}
	return &authn.Basic{
		Username: "_token",
		Password: cgtok,
	}, nil
}
