package platforms

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"time"
)

type GitPlatform string

var (
	GitHubPlatform GitPlatform = "github"
	GitLabPlatform GitPlatform = "gitlab"
)

var successCodes = []int{200, 201, 202, 203, 204}

var (
	ErrBadResponse     = errors.New("request was unsuccessful")
	ErrInvalidPlatform = errors.New("invalid platform")
)

var ValidPlatforms = map[GitPlatform]func(PullRequest) (PRCreator, error){
	GitHubPlatform: NewGitHub,
	GitLabPlatform: NewGitLab,
}

var prTemplate = `{{ $tick := "` + "```" + `" -}}
{{.Description}}

## Changes

<details>

{{ $tick }}diff 

{{ .Diff }}
{{ $tick }}

</details>
`

type PRCreator interface {
	CreatePR(*slog.Logger) error
}

type PullRequest struct {
	Title       string
	Description string
	Diff        string
	Base        string
	Head        string
	Labels      []string
	RepoData
}

type RepoData struct {
	Repo  string
	Owner string
	Token string
}

// Entrypoint to create PR. It's simple but allows us flexibility later.
func CreatePR(p PRCreator, logger *slog.Logger) error {
	return p.CreatePR(logger)
}

func sendRequest(req *http.Request) ([]byte, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(successCodes, resp.StatusCode) {
		return nil, fmt.Errorf("%v status: %d body: %s", ErrBadResponse, resp.StatusCode, string(body))
	}

	return body, nil
}
