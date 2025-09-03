package versioncontrol

import (
	"context"
	"fmt"

	"github.com/chainguard-dev/platform-examples/digestabotctl/platforms"
	"github.com/go-git/go-git/v5"
	"github.com/sigstore/cosign/v2/pkg/providers"
	_ "github.com/sigstore/cosign/v2/pkg/providers/github"
	_ "github.com/sigstore/cosign/v2/pkg/providers/google"
	"github.com/sigstore/gitsign/pkg/fulcio"
	"github.com/sigstore/gitsign/pkg/gitsign"
	"github.com/sigstore/gitsign/pkg/rekor"
	"github.com/sigstore/sigstore/pkg/oauthflow"
)

func init() {
	providers.Register(string(platforms.GitLabPlatform), &GitLab{})
}

func NewSigner(ctx context.Context) (git.Signer, error) {
	if !providers.Enabled(ctx) {
		return nil, fmt.Errorf("no sigstore providers enabled")
	}

	token, err := providers.Provide(ctx, "sigstore")
	if err != nil {
		return nil, err
	}

	fulcio, err := fulcio.NewClient("https://fulcio.sigstore.dev", fulcio.OIDCOptions{
		ClientID: "sigstore",
		Issuer:   "https://oauth2.sigstore.dev/auth",
		TokenGetter: &oauthflow.StaticTokenGetter{
			RawToken: token,
		},
	})
	if err != nil {
		return nil, err
	}

	rekor, err := rekor.NewWithOptions(ctx, "https://rekor.sigstore.dev")
	if err != nil {
		return nil, err
	}

	return gitsign.NewSigner(ctx, fulcio, rekor)
}
