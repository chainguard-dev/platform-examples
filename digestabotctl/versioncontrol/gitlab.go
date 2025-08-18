package versioncontrol

import (
	"context"
	"fmt"
	"os"
)

type GitLab struct{}

func (g *GitLab) Enabled(ctx context.Context) bool {
	return os.Getenv("CI_JOB_TOKEN") != ""
}

func (g *GitLab) Provide(ctx context.Context, audience string) (string, error) {
	token, ok := os.LookupEnv("SIGSTORE_TOKEN")
	if !ok {
		return "", fmt.Errorf("SIGSTORE_TOKEN is not set")
	}

	return token, nil
}
